import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/auth/data/auth_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Signs out (US-A2c): asks the server to revoke the refresh token's
/// family, then forgets the session on this device **whatever the server
/// answered**. The local wipe is what the user asked for; the revocation is
/// best effort (offline, an expired bearer), and a stolen-device scenario is
/// covered by the 30-day sliding expiry and family revocation on reuse.
class SignOut {
  /// Signs [_session] out through [_repository].
  new(this._repository, this._session);

  final AuthRepository _repository;
  final Session _session;

  /// Revokes, then wipes. Never throws an [ApiFailure]: the user ends up
  /// signed out either way, and the router shows the login form.
  Future<void> call() async {
    try {
      // `logout` needs a bearer; after a cold start there is none in memory
      // yet. The interceptor never refreshes for an auth endpoint, so the
      // use case gets one first (a failure here only skips the revocation).
      if (_session.accessToken == null) await _session.refresh();
      final refreshToken = await _session.readRefreshToken();
      if (refreshToken != null) await _repository.logout(refreshToken);
    } on ApiFailure {
      // Best effort by design (see the class doc): the wipe below still runs.
    } finally {
      await _session.end();
    }
  }
}

/// The sign-out use case.
final signOutProvider = Provider<SignOut>(
  (ref) =>
      SignOut(ref.watch(authRepositoryProvider), ref.watch(sessionProvider)),
  name: 'signOutProvider',
);
