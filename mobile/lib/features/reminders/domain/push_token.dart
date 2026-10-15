import 'package:flutter/foundation.dart';

/// This installation's FCM registration token.
///
/// It identifies the install to the push provider, so it is handled like a
/// credential (MODULE-reminders §2.3): [toString] never prints it, and it
/// leaves the app only in the body of `registerDevice`. The generated
/// `DeviceCreate` prints every field in its `toString`, so it is built in
/// exactly one place, right at the call (`ApiRemindersRepository`), and is
/// never stored, logged or reported (`push_token_redaction_test.dart`).
@immutable
final class PushToken {
  /// Wraps the provider's [value].
  const new(this.value);

  /// The raw token; read it only to send it to the API.
  final String value;

  @override
  String toString() => 'PushToken(<redacted>)';

  @override
  bool operator ==(Object other) => other is PushToken && other.value == value;

  @override
  int get hashCode => value.hashCode;
}

/// What the OS lets the app do with notifications.
enum PushPermission {
  /// Notifications allowed (provisional included).
  granted,

  /// The user said no, or blocked them in the system settings.
  denied,

  /// Never asked yet: the next request shows the OS prompt.
  notDetermined,
}
