import 'dart:async';

import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Routes every way out of the form (system back, the app bar's back)
/// through [onLeave], which asks first when there are unsaved changes.
class IncomeLeaveGuard extends StatelessWidget {
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

/// Asks whether to drop what was typed; `true` to leave.
Future<bool> confirmDiscardIncome(BuildContext context) async {
  final l10n = AppLocalizations.of(context);
  final discard = await showDialog<bool>(
    context: context,
    builder: (dialogContext) => AlertDialog(
      title: Text(l10n.incomeDiscardTitle),
      content: Text(l10n.incomeDiscardBody),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(dialogContext).pop(false),
          child: Text(l10n.incomeDiscardKeep),
        ),
        FilledButton(
          onPressed: () => Navigator.of(dialogContext).pop(true),
          child: Text(l10n.incomeDiscardConfirm),
        ),
      ],
    ),
  );
  return discard ?? false;
}
