import 'package:control_contable/core/session/session.dart';
import 'package:flutter/material.dart';

/// Closes every toast of the app's root messenger when [session] ends.
///
/// A record list's Undo toast never auto-dismisses and lives on the root
/// messenger, so without this it would outlive sign-out and offer the next
/// account an Undo of the previous user's records. A token refresh keeps
/// the toasts. Put it in `MaterialApp.builder` (below the root messenger).
class ForgetToastsOnSignOut extends StatefulWidget {
  /// Watches [session] above [child].
  const new({required this.session, required this.child, super.key});

  /// The session whose end closes the toasts.
  final Session session;

  /// The app.
  final Widget child;

  @override
  State<ForgetToastsOnSignOut> createState() => _ForgetToastsOnSignOutState();
}

class _ForgetToastsOnSignOutState extends State<ForgetToastsOnSignOut> {
  @override
  void initState() {
    super.initState();
    widget.session.addListener(_onSessionChanged);
  }

  @override
  void didUpdateWidget(ForgetToastsOnSignOut oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.session == widget.session) return;
    oldWidget.session.removeListener(_onSessionChanged);
    widget.session.addListener(_onSessionChanged);
  }

  @override
  void dispose() {
    widget.session.removeListener(_onSessionChanged);
    super.dispose();
  }

  void _onSessionChanged() {
    if (widget.session.status != SessionStatus.signedOut) return;
    ScaffoldMessenger.maybeOf(context)?.clearSnackBars();
  }

  @override
  Widget build(BuildContext context) => widget.child;
}
