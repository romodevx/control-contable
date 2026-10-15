import 'package:control_contable/core/diagnostics/error_reporter.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/core/session/sign_out_steps.dart';
import 'package:control_contable/features/auth/data/auth_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Signs out (US-A2c): asks the server to revoke the refresh token's
/// family, then forgets the session on this device **whatever the server
/// answered**. The local wipe is what the user asked for; the revocation is
/// best effort (offline, an expired bearer), and a stolen-device scenario is
/// covered by the 30-day sliding expiry and family revocation on reuse.
class SignOut {
  /// Signs [_session] out through [_repository]; a failed revocation is
  /// told to [_reporter]. [beforeRevoke] runs first, with a bearer; the
  /// push device [deviceId] names is removed with the session.
  new(
    this._repository,
    this._session,
    this._reporter, {
    this.beforeRevoke = const [],
    this.deviceId = _noDevice,
  });

  final AuthRepository _repository;
  final Session _session;
  final ErrorReporter _reporter;

  /// What the slices do before the revocation (each failure absorbed).
  final List<SignOutStep> beforeRevoke;

  /// This installation's push device, removed by the server with the
  /// session (a failed lookup is reported and sends none).
  final SignOutDeviceId deviceId;

  static Future<String?> _noDevice() async => null;

  /// Revokes, then wipes. Never throws: the user ends up signed out either
  /// way, and the router shows the login form. A token the wipe could not
  /// remove is reported by `Session.tokenLeftOnDevice`, never by an error.
  Future<void> call() async {
    try {
      // `logout` needs a bearer; after a cold start there is none in memory
      // yet. The interceptor never refreshes for an auth endpoint, so the
      // use case gets one first (a failure here only skips the revocation).
      if (_session.accessToken == null) await _session.refresh();
      for (final step in beforeRevoke) {
        await _absorbed(step);
      }
      final refreshToken = await _session.readRefreshToken();
      if (refreshToken != null) {
        await _repository.logout(refreshToken, deviceId: await _deviceId());
      }
    } on Exception catch (error) {
      // Best effort by design (see the class doc): an `ApiFailure`, or a
      // platform error that may echo the token, is reported by type only;
      // the wipe below still runs (M1-16 FU2).
      _reporter.report('signOut.revoke', error);
    } finally {
      await _session.end();
    }
  }

  Future<String?> _deviceId() async {
    try {
      return await deviceId();
    } on Exception catch (error) {
      _reporter.report('signOut.deviceId', error);
      return null;
    }
  }

  Future<void> _absorbed(SignOutStep step) async {
    try {
      await step();
    } on Exception catch (error) {
      _reporter.report('signOut.beforeRevoke', error);
    }
  }
}

/// The sign-out use case.
final signOutProvider = Provider<SignOut>(
  (ref) => SignOut(
    ref.watch(authRepositoryProvider),
    ref.watch(sessionProvider),
    ref.watch(errorReporterProvider),
    beforeRevoke: ref.watch(signOutStepsProvider),
    deviceId: ref.watch(signOutDeviceIdProvider),
  ),
  name: 'signOutProvider',
);
