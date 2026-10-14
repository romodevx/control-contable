import 'dart:async';

import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/records/stacked_undo.dart';
import 'package:control_contable/features/incomes/application/income_list.dart';
import 'package:control_contable/features/incomes/application/income_records.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// `details[].code` of a restore whose regime no longer covers the date.
const String _regimeNotActive = 'regime_not_active';

/// `details[].code` of a write on an archived taxpayer.
const String _taxpayerArchived = 'taxpayer_archived';

/// Deletes [income] from [query]'s list at once, then offers Undo in the
/// persistent toast (deletes stack). Every account gets the Undo, a
/// read-only one included: restoring gives back the user's own record and
/// creates nothing (contract 0.6.1, MODULE-incomes §2.7).
Future<void> deleteIncome(
  BuildContext context,
  WidgetRef ref,
  IncomeListQuery query,
  Income income,
) async {
  // The toast and its Undo outlive the list: keep what they need now.
  final undo = ref.read(incomesUndoProvider);
  final feedback = _Feedback.of(context, undo);
  final records = ref.read(incomeRecordsProvider);
  try {
    await records.delete(query, income);
  } on ApiFailure catch (failure) {
    feedback.toast(
      _failureMessage(
        failure,
        feedback.l10n,
        fallback: feedback.l10n.incomesDeleteFailed,
      ),
    );
    return;
  }
  undo.push(
    income,
    messenger: feedback.messenger,
    message: feedback.l10n.incomesDeleted,
    undoLabel: feedback.l10n.recordsUndo,
    onUndo: (items) => unawaited(_restore(feedback, records, items)),
  );
}

Future<void> _restore(
  _Feedback feedback,
  IncomeRecords records,
  List<Income> incomes,
) async {
  try {
    await records.restore(incomes);
    feedback.toast(feedback.l10n.incomesRestored(incomes.length));
  } on ApiFailure catch (failure) {
    feedback.toast(
      _failureMessage(
        failure,
        feedback.l10n,
        fallback: feedback.l10n.incomesRestoreFailed,
      ),
    );
  }
}

/// Why a delete or restore was refused; [fallback] when no rule says.
String _failureMessage(
  ApiFailure failure,
  AppLocalizations l10n, {
  required String fallback,
}) => switch (failure) {
  ApiRejected(:final details)
      when details.any((d) => d.code == _regimeNotActive) =>
    l10n.incomesRestoreRegimeEnded,
  ApiRejected(:final details)
      when details.any((d) => d.code == _taxpayerArchived) =>
    l10n.settingsArchivedWrite,
  _ => fallback,
};

/// What a delete's follow-up needs, read while the list is mounted.
final class _Feedback {
  new of(BuildContext context, this.undo)
    : messenger = ScaffoldMessenger.of(context),
      l10n = AppLocalizations.of(context);

  final ScaffoldMessengerState messenger;
  final AppLocalizations l10n;
  final StackedUndo<Income> undo;

  /// A plain toast; it never closes a pending Undo (PR #85 follow-up).
  void toast(String message) => undo.notify(messenger, message);
}
