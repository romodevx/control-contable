import 'dart:async';

import 'package:control_contable/api/generated/models/expense.dart';
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/core/access/paywall_sheet.dart';
import 'package:control_contable/core/records/record_save_toast.dart';
import 'package:control_contable/core/records/stacked_undo.dart';
import 'package:control_contable/features/expenses/presentation/expense_copy.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Opens the paywall over the form (its data stays): the expenses trial
/// cap names the cap; otherwise the read-only message.
void showExpensePaywall(BuildContext context, {required bool isTrialCap}) {
  final l10n = AppLocalizations.of(context);
  unawaited(
    showPaywallSheet(
      context,
      title: isTrialCap
          ? l10n.expenseTrialCapTitle
          : l10n.onboardingPaywallTitle,
      body: isTrialCap ? l10n.expenseTrialCapBody : l10n.paywallWriteBody,
    ),
  );
}

/// Confirms a save ([updated] for an edit) on [messenger] (see
/// [showRecordSavedToast]): warnings never block, and a pending Undo of
/// [undo] is never closed by it.
void showExpenseSavedToast(
  ScaffoldMessengerState messenger,
  AppLocalizations l10n,
  List<Warning> warnings, {
  required StackedUndo<Expense> undo,
  bool updated = false,
  bool overForm = false,
}) {
  final count = warnings.length;
  showRecordSavedToast(
    messenger,
    heading: switch ((count > 0, updated)) {
      (true, false) => l10n.expenseSavedWithWarnings(count),
      (true, true) => l10n.expenseUpdatedWithWarnings(count),
      (false, false) => l10n.expenseSaved,
      (false, true) => l10n.expenseUpdated,
    },
    notes: [for (final w in warnings) describeExpenseWarning(w, l10n)],
    undo: undo,
    overForm: overForm,
  );
}
