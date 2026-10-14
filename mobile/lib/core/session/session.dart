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

/// How many times a secure-store write or delete is tried before the
/// session falls back (see [Session]).
const int storageAttempts = 3;

/// The waits between those tries: one fewer than [storageAttempts].
const List<Duration> defaultStorageRetryDelays = [
  Duration(milliseconds: 200),
  Duration(seconds: 1),
];

/// The signed-in session: the access token in memory only, the rotating
/// refresh token in a [RefreshTokenStore] (secure storage). Notifies on
/// every [status] change, which re-runs the router's guards.
///
/// [refresh] is single-flight: every caller during a rotation waits on the
/// same one, so concurrent 401s spend the refresh token exactly once (the
/// server treats a second use as reuse and revokes the family).
///
/// The secure store can fail (a locked Keychain, a Keystore hiccup); every
/// write and delete is retried, and the session never throws the store's
/// error (some echo the value they failed on):
/// - a refresh token that could not be saved stays in memory, the session
///   stays signed in, and the next rotation sends it (never the spent one
///   still on disk, which the server would treat as reuse) and saves again;
/// - a refresh token that could not be deleted at [end] makes the session
///   wipe the whole store; when even that fails, [tokenLeftOnDevice] says so.
class Session extends ChangeNotifier {
  /// A session kept in `store`, rotated through `rotate`. A failed store
  /// call is retried after each of `storageRetryDelays`.
  new({
    required this._store,
    required this._rotate,
    this._restoreTimeout = sessionRestoreTimeout,
    this._storageRetryDelays = defaultStorageRetryDelays,
  });

  final RefreshTokenStore _store;
  final TokenRotation _rotate;
  final Duration _restoreTimeout;
  final List<Duration> _storageRetryDelays;

  SessionStatus _status = SessionStatus.restoring;
  String? _accessToken;
  String? _unsavedRefreshToken;
  bool _tokenLeftOnDevice = false;
  Future<String?>? _refreshing;
  bool _isDisposed = false;

  /// Bumped by every [start] and [end]: work that began under an older
  /// generation (a rotation, a retried store call) must not touch the
  /// newer session.
  int _generation = 0;

  /// Whether someone is signed in.
  SessionStatus get status => _status;

  /// Whether a refresh token is kept.
  bool get isSignedIn => _status == SessionStatus.signedIn;

  /// The current access token; `null` until a sign-in or a refresh.
  String? get accessToken => _accessToken;

  /// Whether the last sign-out could not remove the refresh token from
  /// this device, even by wiping the store. The next sign-in overwrites it.
  bool get tokenLeftOnDevice => _tokenLeftOnDevice;

  /// Reads the stored refresh token once at boot. An unreadable store
  /// counts as no session: the user signs in again.
  ///
  /// A read slower than the restore timeout signs out for now (the splash
  /// stays bounded), but its late answer still restores the session: a
  /// slow Keystore on a cold start must not look like a sign-out (M1-15
  /// QA D1). A sign-in or sign-out in between wins over that answer.
  Future<void> restore() async {
    if (_status != SessionStatus.restoring) return;
    final generation = _generation;
    final pending = _store.read();
    String? stored;
    try {
      stored = await pending.timeout(_restoreTimeout);
    } on TimeoutException {
      _settleRestore(generation, null);
      unawaited(_restoreLate(pending, generation));
      return;
    } on Exception {
      stored = null;
    }
    _settleRestore(generation, stored);
  }

  void _settleRestore(int generation, String? stored) {
    if (generation != _generation || _status != SessionStatus.restoring) {
      return; // A sign-in or sign-out finished first and is the truth.
    }
    _setStatus(
      stored == null ? SessionStatus.signedOut : SessionStatus.signedIn,
    );
  }

  Future<void> _restoreLate(Future<String?> pending, int generation) async {
    final String? stored;
    try {
      stored = await pending;
    } on Exception {
      return; // Still unreadable: the user signs in again.
    }
    if (stored == null || generation != _generation) return;
    if (_status == SessionStatus.signedOut) _setStatus(SessionStatus.signedIn);
  }

  /// Starts the session [tokens] belong to (`login`, `verifyEmail`). The
  /// refresh token is saved (or kept in memory) before the session counts
  /// as signed in.
  Future<void> start(TokenPair tokens) async {
    final generation = ++_generation;
    await _keep(tokens.refreshToken, generation);
    _accessToken = tokens.accessToken;
    _setStatus(SessionStatus.signedIn);
  }

  /// The refresh token (for `logout`): the unsaved one when the store
  /// lags, else the stored one, or `null`.
  Future<String?> readRefreshToken() async =>
      _unsavedRefreshToken ?? await _store.read();

  /// A new access token, rotating the refresh token once for every caller
  /// waiting at the same time. `null` when there is none: the server
  /// refused the refresh token (the session is then ended) or could not be
  /// reached (the session is kept; the caller's own error stands).
  Future<String?> refresh() =>
      _refreshing ??= _rotateOnce().whenComplete(() => _refreshing = null);

  /// Ends the session locally: both tokens are forgotten. Never throws.
  Future<void> end() async {
    final generation = ++_generation;
    _accessToken = null;
    _unsavedRefreshToken = null;
    _setStatus(SessionStatus.signedOut);
    final removed =
        await _retrying(_store.delete, generation) ||
        await _retrying(_store.wipeAll, generation);
    if (removed || generation != _generation) return;
    _tokenLeftOnDevice = true;
    _notify();
  }

  Future<String?> _rotateOnce() async {
    final generation = _generation;
    final String? current;
    try {
      current = _unsavedRefreshToken ?? await _store.read();
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
    await _keep(pair.refreshToken, generation);
    if (generation != _generation) return null;
    _accessToken = pair.accessToken;
    return pair.accessToken;
  }

  /// Holds [refreshToken] in memory at once (it is the only live one now),
  /// then saves it; it leaves memory only once the store has it.
  Future<void> _keep(String refreshToken, int generation) async {
    _unsavedRefreshToken = refreshToken;
    final saved = await _retrying(() => _store.write(refreshToken), generation);
    if (!saved || generation != _generation) return;
    _unsavedRefreshToken = null;
    _tokenLeftOnDevice = false; // The key now holds the live token.
  }

  /// Runs [storeCall], retrying after each of the configured delays while
  /// [generation] is current. Whether it succeeded. The store's exception
  /// is dropped on purpose: it may carry the token.
  Future<bool> _retrying(
    Future<void> Function() storeCall,
    int generation,
  ) async {
    for (var attempt = 0; generation == _generation; attempt++) {
      try {
        await storeCall();
        return true;
      } on Exception {
        if (attempt >= _storageRetryDelays.length) return false;
      }
      await Future<void>.delayed(_storageRetryDelays[attempt]);
    }
    return false;
  }

  Future<String?> _endIfCurrent(int generation) async {
    if (generation == _generation) await end();
    return null;
  }

  void _setStatus(SessionStatus status) {
    if (_status == status) return;
    _status = status;
    _notify();
  }

  void _notify() {
    if (!_isDisposed) notifyListeners();
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
