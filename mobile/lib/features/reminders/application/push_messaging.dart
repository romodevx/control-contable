import 'package:control_contable/features/reminders/domain/push_token.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The push provider as the reminders slice needs it (port). The Firebase
/// adapter (`FirebasePushMessaging`) implements it when Firebase started;
/// [UnavailablePushMessaging] otherwise, so a build without the Firebase
/// config files (OB9) runs, and the screen says push is not available.
abstract interface class PushMessaging {
  /// Whether push works in this build (Firebase configured and started).
  bool get isAvailable;

  /// The current OS permission, without prompting.
  Future<PushPermission> permission();

  /// Shows the OS prompt when never asked; else answers the current state.
  Future<PushPermission> requestPermission();

  /// This installation's token, or `null` when the provider has none.
  Future<PushToken?> token();

  /// A new token each time the provider rotates it.
  Stream<PushToken> get tokenRefreshes;

  /// The deep link of the notification that launched the app, if any.
  Future<String?> initialLink();

  /// The deep link of each notification tapped while the app runs.
  Stream<String> get openedLinks;
}

/// Raised when push cannot work on this device right now (no Firebase
/// config, no token from the provider).
final class PushUnavailable implements Exception {
  /// The failure.
  const new();
}

/// [PushMessaging] of a build where Firebase did not start: nothing to ask,
/// no token, no taps.
class UnavailablePushMessaging implements PushMessaging {
  /// The stand-in.
  const new();

  @override
  bool get isAvailable => false;

  @override
  Future<PushPermission> permission() async => PushPermission.denied;

  @override
  Future<PushPermission> requestPermission() async => PushPermission.denied;

  @override
  Future<PushToken?> token() async => null;

  @override
  Stream<PushToken> get tokenRefreshes => const Stream.empty();

  @override
  Future<String?> initialLink() async => null;

  @override
  Stream<String> get openedLinks => const Stream.empty();
}

/// The push provider. `bootstrap` overrides it with the Firebase adapter
/// when Firebase starts; tests with a fake.
final pushMessagingProvider = Provider<PushMessaging>(
  (ref) => const UnavailablePushMessaging(),
  name: 'pushMessagingProvider',
);
