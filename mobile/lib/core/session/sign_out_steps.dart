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
