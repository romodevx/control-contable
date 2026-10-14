import 'dart:async';

import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Routes every way out of a form page (system back, the app bar's back)
/// through [onLeave], which asks first when there are unsaved changes.
class FormLeaveGuard extends StatelessWidget {
  /// Guards [child].
  const new({required this.onLeave, required this.child, super.key});

  /// Leaves, asking first when needed.
  final Future<void> Function() onLeave;

  /// The page.
  final Widget child;

  @override
  Widget build(BuildContext context) => PopScope(
    canPop: false,
    onPopInvokedWithResult: (didPop, _) {
      if (!didPop) unawaited(onLeave());
    },
    child: child,
  );
}

/// Asks whether to drop what was typed in the record named by [title]
/// ("¿Descartar este gasto?"); `true` to leave.
Future<bool> confirmDiscardRecord(
  BuildContext context, {
  required String title,
}) async {
  final l10n = AppLocalizations.of(context);
  final discard = await showDialog<bool>(
    context: context,
    builder: (dialogContext) => AlertDialog(
      title: Text(title),
      content: Text(l10n.recordDiscardBody),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(dialogContext).pop(false),
          child: Text(l10n.recordDiscardKeep),
        ),
        FilledButton(
          onPressed: () => Navigator.of(dialogContext).pop(true),
          child: Text(l10n.recordDiscardConfirm),
        ),
      ],
    ),
  );
  return discard ?? false;
}
