import 'dart:async';

import 'package:control_contable/core/time/clock.dart';
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

  /// The chosen password; `null` once forgotten (after [signUpIdleTimeout]
  /// without activity) and the code screen must ask for it again.
  final String? password;

  /// The optional display name.
  final String? name;

  /// The same draft without the password.
  SignUpDraft withoutPassword() => SignUpDraft(email: email, name: name);
}

/// How long the password waits without activity on the code step (the
/// code itself lives 15 minutes).
const Duration signUpIdleTimeout = Duration(minutes: 15);

/// The sign-up waiting for its code. Held in memory only: never persisted,
/// never in a route, so process death drops it. The password survives a
/// trip to the mail app; it is forgotten after [signUpIdleTimeout] without
/// activity (on the `clockProvider`, checked by a timer and on every
/// resume, since a suspended app's timers may not fire), and the whole
/// draft on success or when the person goes back to the form.
class PendingSignUp extends Notifier<SignUpDraft?> {
  Timer? _idleTimer;
  DateTime? _lastActivity;

  @override
  SignUpDraft? build() {
    final lifecycle = AppLifecycleListener(onResume: _forgetPasswordIfIdle);
    ref.onDispose(() {
      lifecycle.dispose();
      _idleTimer?.cancel();
    });
    return null;
  }

  /// Remembers [draft] as the sign-up in progress.
  void start(SignUpDraft draft) {
    state = draft;
    touch();
  }

  /// Records activity on the code step: the idle timeout starts over.
  void touch() {
    if (state?.password == null) return;
    _lastActivity = ref.read(clockProvider)();
    _idleTimer?.cancel();
    _idleTimer = Timer(signUpIdleTimeout, _forgetPasswordIfIdle);
  }

  /// Forgets the whole draft (account created, or back to the form).
  void clear() {
    _idleTimer?.cancel();
    state = null;
  }

  void _forgetPasswordIfIdle() {
    final draft = state;
    final lastActivity = _lastActivity;
    if (draft?.password == null || lastActivity == null) return;
    final idle = ref.read(clockProvider)().difference(lastActivity);
    if (idle < signUpIdleTimeout) return;
    _idleTimer?.cancel();
    state = draft!.withoutPassword();
  }
}

/// The sign-up waiting for its code, or `null`.
final pendingSignUpProvider = NotifierProvider<PendingSignUp, SignUpDraft?>(
  PendingSignUp.new,
  name: 'pendingSignUpProvider',
);
