import 'dart:async';

import 'package:control_contable/api/generated/models/salary_statement.dart';
import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/annual/application/annual_input_writes.dart';
import 'package:control_contable/features/annual/presentation/annual_write_feedback.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// One constancia: its name, its gravados and ISR withheld; the row opens
/// the edit sheet, the button deletes it after asking.
class SalaryStatementTile extends ConsumerWidget {
  /// The row of [statement], called [name].
  const new({
    required this.statement,
    required this.name,
    required this.onEdit,
    super.key,
  });

  /// The constancia.
  final SalaryStatement statement;

  /// Its alias, or "Constancia n".
  final String name;

  /// Opens the edit sheet.
  final VoidCallback onEdit;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    return ListTile(
      contentPadding: EdgeInsets.zero,
      title: Text(name),
      subtitle: Text(
        '${l10n.annualSalaryAccumulable}: '
        '${formatPesos(statement.accumulableCents)} · '
        '${l10n.annualSalaryIsrWithheld}: '
        '${formatPesos(statement.isrWithheldCents)}',
      ),
      onTap: onEdit,
      trailing: IconButton(
        tooltip: l10n.annualSalaryDeleteTooltip(name),
        icon: const Icon(Icons.delete_outline),
        onPressed: () => unawaited(_delete(context, ref, l10n)),
      ),
    );
  }

  Future<void> _delete(
    BuildContext context,
    WidgetRef ref,
    AppLocalizations l10n,
  ) async {
    final messenger = ScaffoldMessenger.of(context);
    final writes = ref.read(annualInputWritesProvider);
    if (!await _confirm(context, l10n)) return;
    String message;
    try {
      await writes.deleteStatement(statement);
      message = l10n.annualSalaryDeleted;
    } on ApiFailure catch (failure) {
      message = refusalText(
        failure,
        l10n,
        fallback: l10n.annualSalaryDeleteFailed,
      );
    }
    messenger
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(content: Text(message)));
  }

  /// The delete cannot be undone (hard delete): ask first.
  Future<bool> _confirm(BuildContext context, AppLocalizations l10n) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: Text(l10n.annualSalaryDeleteTitle(name)),
        content: Text(l10n.annualSalaryDeleteBody),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(dialogContext).pop(false),
            child: Text(l10n.annualSalaryDeleteKeep),
          ),
          FilledButton(
            onPressed: () => Navigator.of(dialogContext).pop(true),
            child: Text(l10n.annualSalaryDeleteConfirm),
          ),
        ],
      ),
    );
    return confirmed ?? false;
  }
}
