import 'dart:async';

import 'package:control_contable/api/generated/clients/auth_client.dart';
import 'package:control_contable/api/generated/models/refresh_request.dart';
import 'package:control_contable/api/generated/models/token_pair.dart';
import 'package:control_contable/core/network/api_dio.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/session/refresh_token_store.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Whether someone is signed in.
enum SessionStatus {
  /// The stored refresh token has not been read yet (cold start).
  restoring,

  /// A refresh token is kept; the access token may still need a refresh.
  signedIn,

  /// No session: the user must sign in.
  signedOut,
}

/// Exchanges a refresh token for a new pair (`POST /v1/auth/refresh`);
/// throws an [ApiFailure] when refused or unreachable.
typedef TokenRotation = Future<TokenPair> Function(String refreshToken);

/// The longest a cold start waits on the secure store before it treats the
/// session as absent (a locked Keychain must not hold the splash forever).
const Duration sessionRestoreTimeout = Duration(seconds: 5);

/// The signed-in session: the access token in memory only, the rotating
/// refresh token in a [RefreshTokenStore] (secure storage). Notifies on
/// every [status] change, which re-runs the router's guards.
///
/// [refresh] is single-flight: every caller during a rotation waits on the
/// same one, so concurrent 401s spend the refresh token exactly once (the
/// server treats a second use as reuse and revokes the family). The newly
/// returned refresh token is always stored before anyone retries.
class Session extends ChangeNotifier {
  /// A session kept in `store`, rotated through `rotate`.
  new({
    required this._store,
    required this._rotate,
    this._restoreTimeout = sessionRestoreTimeout,
  });

  final RefreshTokenStore _store;
  final TokenRotation _rotate;
  final Duration _restoreTimeout;

  SessionStatus _status = SessionStatus.restoring;
  String? _accessToken;
  Future<String?>? _refreshing;
  bool _isDisposed = false;

  /// Bumped by every [start] and [end]: a rotation that began under an
  /// older generation must not write its tokens back.
  int _generation = 0;

  /// Whether someone is signed in.
  SessionStatus get status => _status;

  /// Whether a refresh token is kept.
  bool get isSignedIn => _status == SessionStatus.signedIn;

  /// The current access token; `null` until a sign-in or a refresh.
  String? get accessToken => _accessToken;

  /// Reads the stored refresh token once at boot. An unreadable store
  /// counts as no session: the user signs in again.
  Future<void> restore() async {
    if (_status != SessionStatus.restoring) return;
    final generation = _generation;
    String? stored;
    try {
      stored = await _store.read().timeout(_restoreTimeout);
    } on Exception {
      stored = null;
    }
    if (generation != _generation || _status != SessionStatus.restoring) {
      return; // A sign-in or sign-out finished first and is the truth.
    }
    _setStatus(
      stored == null ? SessionStatus.signedOut : SessionStatus.signedIn,
    );
  }

  /// Starts the session [tokens] belong to (`login`, `verifyEmail`). The
  /// refresh token is stored before the session counts as signed in.
  Future<void> start(TokenPair tokens) async {
    _generation++;
    await _store.write(tokens.refreshToken);
    _accessToken = tokens.accessToken;
    _setStatus(SessionStatus.signedIn);
  }

  /// The stored refresh token (for `logout`), or `null`.
  Future<String?> readRefreshToken() => _store.read();

  /// A new access token, rotating the refresh token once for every caller
  /// waiting at the same time. `null` when there is none: the server
  /// refused the refresh token (the session is then ended) or could not be
  /// reached (the session is kept; the caller's own error stands).
  Future<String?> refresh() =>
      _refreshing ??= _rotateOnce().whenComplete(() => _refreshing = null);

  /// Ends the session locally: both tokens are forgotten.
  Future<void> end() async {
    _generation++;
    _accessToken = null;
    _setStatus(SessionStatus.signedOut);
    await _store.delete();
  }

  Future<String?> _rotateOnce() async {
    final generation = _generation;
    final String? current;
    try {
      current = await _store.read();
    } on Exception {
      return null; // The store failed this time; the session stands.
    }
    if (current == null) {
      await end();
      return null;
    }
    final TokenPair pair;
    try {
      pair = await _rotate(current);
    } on ApiUnauthorized {
      return await _endIfCurrent(generation);
    } on ApiValidationFailure {
      return await _endIfCurrent(generation);
    } on ApiFailure {
      return null; // Transient (offline, 5xx, 429): try again later.
    }
    if (generation != _generation) return null;
    await _store.write(pair.refreshToken);
    _accessToken = pair.accessToken;
    return pair.accessToken;
  }

  Future<String?> _endIfCurrent(int generation) async {
    if (generation == _generation) await end();
    return null;
  }

  void _setStatus(SessionStatus status) {
    if (_isDisposed || _status == status) return;
    _status = status;
    notifyListeners();
  }

  @override
  void dispose() {
    _isDisposed = true;
    super.dispose();
  }
}

/// How the session rotates its refresh token: `refreshSession` of the
/// generated client on the bare client (no auth interceptor, so a refused
/// refresh never triggers another). Tests override it.
final tokenRotationProvider = Provider<TokenRotation>((ref) {
  final client = AuthClient(ref.watch(bareDioProvider));
  return (refreshToken) => guardApi(
    () =>
        client.refreshSession(body: RefreshRequest(refreshToken: refreshToken)),
  );
}, name: 'tokenRotationProvider');

/// The app's session; reading it starts the restore from secure storage.
final sessionProvider = Provider<Session>((ref) {
  final session = Session(
    store: ref.watch(refreshTokenStoreProvider),
    rotate: ref.watch(tokenRotationProvider),
  );
  ref.onDispose(session.dispose);
  unawaited(session.restore());
  return session;
}, name: 'sessionProvider');
