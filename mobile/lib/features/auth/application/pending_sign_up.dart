import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The email of the sign-up waiting for its code, held in memory between
/// the register and verify screens. Kept out of the route so the address
/// never sits in a URL, a log line or restored navigation state.
class PendingSignUp extends Notifier<String?> {
  @override
  String? build() => null;

  /// Remembers [email] as the sign-up in progress.
  // A verb reads better than a setter at the call site.
  // ignore: use_setters_to_change_properties
  void start(String email) => state = email;

  /// Forgets it (the account was created).
  void clear() => state = null;
}

/// The sign-up waiting for its code, or `null`.
final pendingSignUpProvider = NotifierProvider<PendingSignUp, String?>(
  PendingSignUp.new,
  name: 'pendingSignUpProvider',
);
