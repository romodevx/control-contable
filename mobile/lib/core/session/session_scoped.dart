import 'package:control_contable/core/session/session.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Ties the provider being built through [ref] to the signed-in user: when
/// the session ends it is rebuilt from scratch, so what it remembered "in
/// this session" (a last choice, an Undo stack) never reaches the next
/// account on the device. A token refresh keeps it. Call it from `build`.
void forgetOnSignOut(Ref ref) {
  final session = ref.watch(sessionProvider);
  void onChanged() {
    if (session.status == SessionStatus.signedOut) ref.invalidateSelf();
  }

  session.addListener(onChanged);
  ref.onDispose(() => session.removeListener(onChanged));
}
