import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Work a module must do while the user is still signed in, right before
/// the session is revoked (unregister this device's push token, M9-09).
/// It runs with a bearer; a failure is reported and never stops the
/// sign-out.
typedef SignOutStep = Future<void> Function();

/// The steps each slice adds, wired at the app root so `auth` never
/// imports another feature. None by default.
final signOutStepsProvider = Provider<List<SignOutStep>>(
  (ref) => const [],
  name: 'signOutStepsProvider',
);

/// This installation's push device id for the server to remove with the
/// session (`logout`'s `device_id`, M9 security gate F1), or null for
/// none. A failure is reported and the sign-out goes on without it.
typedef SignOutDeviceId = Future<String?> Function();

/// The device id lookup, wired at the app root by the slice that owns push
/// so `auth` never imports it. None by default.
final signOutDeviceIdProvider = Provider<SignOutDeviceId>(
  (ref) =>
      () async => null,
  name: 'signOutDeviceIdProvider',
);
