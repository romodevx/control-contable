import 'dart:async';

import 'package:control_contable/api/generated/models/expense.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/records/record_list.dart';
import 'package:control_contable/core/records/stacked_undo.dart';
import 'package:control_contable/features/expenses/application/expense_records.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// `details[].code` of a restore whose regime no longer covers the date.
const String _regimeNotActive = 'regime_not_active';

/// `details[].code` of a write on an archived taxpayer.
const String _taxpayerArchived = 'taxpayer_archived';

/// Deletes [expense] from [query]'s list at once, then offers Undo in the
/// persistent toast (deletes stack). Every account gets the Undo, a
/// read-only one included: restoring never answers `402` (contract 0.6.1).
Future<void> deleteExpense(
  BuildContext context,
  WidgetRef ref,
  RecordListQuery query,
  Expense expense,
) async {
  // The toast and its Undo outlive the list: keep what they need now.
  final undo = ref.read(expensesUndoProvider);
  final feedback = _Feedback.of(context, undo);
  final records = ref.read(expenseRecordsProvider);
  try {
    await records.delete(query, expense);
  } on ApiFailure catch (failure) {
    feedback.toast(
      _refusalMessage(failure, feedback.l10n) ??
          feedback.l10n.expensesDeleteFailed,
    );
    return;
  }
  undo.push(
    expense,
    messenger: feedback.messenger,
    message: feedback.l10n.expensesDeleted,
    undoLabel: feedback.l10n.recordsUndo,
    onUndo: (items) => unawaited(_restore(feedback, records, items)),
  );
}

/// Restores every one of [expenses]; the toast says how many came back
/// and, when some could not, why.
Future<void> _restore(
  _Feedback feedback,
  ExpenseRecords records,
  List<Expense> expenses,
) async {
  final l10n = feedback.l10n;
  final report = await records.restore(expenses);
  if (report.refusals.isEmpty) {
    feedback.toast(l10n.expensesRestored(report.restored));
    return;
  }
  final reason =
      _refusalMessage(report.refusals.first, l10n) ??
      l10n.expensesRestoreFailed;
  feedback.toast(
    expenses.length == 1
        ? reason
        : '${l10n.expensesRestoredSome(report.restored, expenses.length)} '
              '$reason',
  );
}

/// Why a delete or restore was refused, when a rule says.
String? _refusalMessage(ApiFailure failure, AppLocalizations l10n) =>
    switch (failure) {
      ApiRejected(:final details)
          when details.any((d) => d.code == _regimeNotActive) =>
        l10n.expensesRestoreRegimeEnded,
      ApiRejected(:final details)
          when details.any((d) => d.code == _taxpayerArchived) =>
        l10n.settingsArchivedWrite,
      _ => null,
    };

/// What a delete's follow-up needs, read while the list is mounted.
final class _Feedback {
  new of(BuildContext context, this.undo)
    : messenger = ScaffoldMessenger.of(context),
      l10n = AppLocalizations.of(context);

  final ScaffoldMessengerState messenger;
  final AppLocalizations l10n;
  final StackedUndo<Expense> undo;

  /// A plain toast; it never closes a pending Undo (PR #85 follow-up).
  void toast(String message) => undo.notify(messenger, message);
}
