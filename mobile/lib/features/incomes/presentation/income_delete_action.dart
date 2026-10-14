import 'dart:async';

import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/core/access/paywall_sheet.dart';
import 'package:control_contable/core/network/api_failure.dart';
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
/// persistent toast (deletes stack). Once the session knows the account is
/// read-only, the delete is confirmed first and offers no Undo: the server
/// would refuse the restore (MODULE-incomes §2.10).
Future<void> deleteIncome(
  BuildContext context,
  WidgetRef ref,
  IncomeListQuery query,
  Income income,
) async {
  // The toast and its Undo outlive the list: keep what they need now.
  final feedback = _Feedback.of(context);
  final records = ref.read(incomeRecordsProvider);
  final undo = ref.read(incomesUndoProvider);
  final oneWay = records.knownReadOnly;
  if (oneWay && !await _confirmOneWay(context)) return;
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
  if (oneWay) {
    feedback.toast(feedback.l10n.incomesDeleted(1));
    return;
  }
  undo.push(
    income,
    messenger: feedback.messenger,
    message: feedback.l10n.incomesDeleted,
    undoLabel: feedback.l10n.recordsUndo,
    onUndo: (items) => unawaited(_restore(context, feedback, records, items)),
  );
}

Future<void> _restore(
  BuildContext context,
  _Feedback feedback,
  IncomeRecords records,
  List<Income> incomes,
) async {
  try {
    await records.restore(incomes);
    feedback.toast(feedback.l10n.incomesRestored(incomes.length));
  } on ApiEntitlementRequired {
    final l10n = feedback.l10n;
    if (!context.mounted) return feedback.toast(l10n.paywallWriteBody);
    unawaited(
      showPaywallSheet(
        context,
        title: l10n.onboardingPaywallTitle,
        body: l10n.paywallWriteBody,
      ),
    );
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

/// Asks a read-only account to confirm a delete it cannot undo.
Future<bool> _confirmOneWay(BuildContext context) async {
  final l10n = AppLocalizations.of(context);
  final confirmed = await showDialog<bool>(
    context: context,
    builder: (dialogContext) => AlertDialog(
      title: Text(l10n.incomesDeleteOneWayTitle),
      content: Text(l10n.incomesDeleteOneWayBody),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(dialogContext).pop(false),
          child: Text(l10n.formCancel),
        ),
        FilledButton(
          onPressed: () => Navigator.of(dialogContext).pop(true),
          child: Text(l10n.recordsDelete),
        ),
      ],
    ),
  );
  return confirmed ?? false;
}

/// What a delete's follow-up needs, read while the list is mounted.
final class _Feedback {
  new of(BuildContext context)
    : messenger = ScaffoldMessenger.of(context),
      l10n = AppLocalizations.of(context);

  final ScaffoldMessengerState messenger;
  final AppLocalizations l10n;

  /// A plain toast (it replaces the current one).
  void toast(String message) => messenger
    ..hideCurrentSnackBar()
    ..showSnackBar(SnackBar(content: Text(message)));
}
