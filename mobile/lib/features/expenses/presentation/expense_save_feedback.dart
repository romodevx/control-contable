import 'dart:async';

import 'package:control_contable/api/generated/models/expense.dart';
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/core/access/paywall_sheet.dart';
import 'package:control_contable/core/forms/form_scope.dart';
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

/// Confirms a save ([updated] for an edit) on [messenger]. Warnings never
/// block: they are read in the toast, which then stays until closed so
/// they are not missed. A pending Undo of [undo] is never closed by it
/// (the confirmation waits behind it). [overForm]: the form stays open
/// (*Guardar y crear otro*), so the toast floats above its footer.
void showExpenseSavedToast(
  ScaffoldMessengerState messenger,
  AppLocalizations l10n,
  List<Warning> warnings, {
  required StackedUndo<Expense> undo,
  bool updated = false,
  bool overForm = false,
}) {
  final hasWarnings = warnings.isNotEmpty;
  final notes = [for (final w in warnings) describeExpenseWarning(w, l10n)];
  if (undo.pending.isEmpty) messenger.hideCurrentSnackBar();
  messenger.showSnackBar(
    SnackBar(
      content: Text(switch ((hasWarnings, updated)) {
        (true, false) => '${l10n.expenseSavedWithWarnings} ${notes.join(' ')}',
        (true, true) => '${l10n.expenseUpdatedWithWarnings} ${notes.join(' ')}',
        (false, false) => l10n.expenseSaved,
        (false, true) => l10n.expenseUpdated,
      }),
      persist: hasWarnings,
      showCloseIcon: hasWarnings,
      behavior: overForm ? SnackBarBehavior.floating : null,
      margin: overForm ? toastClearOfFormFooter : null,
    ),
  );
}
