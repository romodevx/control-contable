import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// What the sign-up form collected for the code step. Since D8 the password
/// (and the optional name) travel with the code in `verifyEmail`, never in
/// `register`, so they wait here, in memory only.
@immutable
class SignUpDraft {
  /// A draft for [email], with the [password] and [name] chosen.
  const new({required this.email, this.password, this.name});

  /// The address the code was sent to.
  final String email;

  /// The chosen password; `null` once forgotten (the app went to the
  /// background) and the code screen must ask for it again.
  final String? password;

  /// The optional display name.
  final String? name;

  /// The same draft without the password.
  SignUpDraft withoutPassword() => SignUpDraft(email: email, name: name);
}

/// The sign-up waiting for its code. Held in memory only: never persisted,
/// never in a route. The password is forgotten as soon as the app goes to
/// the background, and the whole draft on success or when the person goes
/// back to the form.
class PendingSignUp extends Notifier<SignUpDraft?> {
  @override
  SignUpDraft? build() {
    final lifecycle = AppLifecycleListener(onPause: _forgetPassword);
    ref.onDispose(lifecycle.dispose);
    return null;
  }

  /// Remembers [draft] as the sign-up in progress.
  // A verb reads better than a setter at the call site.
  // ignore: use_setters_to_change_properties
  void start(SignUpDraft draft) => state = draft;

  /// Forgets the whole draft (account created, or back to the form).
  void clear() => state = null;

  void _forgetPassword() {
    final draft = state;
    if (draft?.password != null) state = draft!.withoutPassword();
  }
}

/// The sign-up waiting for its code, or `null`.
final pendingSignUpProvider = NotifierProvider<PendingSignUp, SignUpDraft?>(
  PendingSignUp.new,
  name: 'pendingSignUpProvider',
);
