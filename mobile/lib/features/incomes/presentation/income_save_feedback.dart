import 'dart:async';

import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/core/access/paywall_sheet.dart';
import 'package:control_contable/features/incomes/application/income_save.dart';
import 'package:control_contable/features/incomes/presentation/income_form_copy.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Opens the paywall over the form (its data stays): the incomes trial cap
/// names the cap; otherwise the read-only message.
void showIncomePaywall(BuildContext context, {required bool isTrialCap}) {
  final l10n = AppLocalizations.of(context);
  unawaited(
    showPaywallSheet(
      context,
      title: isTrialCap
          ? l10n.incomeTrialCapTitle
          : l10n.onboardingPaywallTitle,
      body: isTrialCap ? l10n.incomeTrialCapBody : l10n.paywallWriteBody,
    ),
  );
}

/// The form-level message of a save that did not go through.
String incomeSaveNoticeText(IncomeSaveNotice notice, AppLocalizations l10n) =>
    switch (notice) {
      IncomeSaveNotice.unreachable => l10n.incomeSaveUnreachable,
      IncomeSaveNotice.idReused => l10n.incomeSaveIdReused,
      IncomeSaveNotice.failed => l10n.incomeSaveFailed,
    };

/// Confirms a save on [messenger]. Warnings never block: they are read in
/// the toast, which then stays until closed so they are not missed.
void showIncomeSavedToast(
  ScaffoldMessengerState messenger,
  AppLocalizations l10n,
  List<Warning> warnings,
) {
  final hasWarnings = warnings.isNotEmpty;
  final notes = [for (final w in warnings) describeIncomeWarning(w, l10n)];
  messenger
    ..hideCurrentSnackBar()
    ..showSnackBar(
      SnackBar(
        content: Text(
          hasWarnings
              ? '${l10n.incomeSavedWithWarnings} ${notes.join(' ')}'
              : l10n.incomeSaved,
        ),
        persist: hasWarnings,
        showCloseIcon: hasWarnings,
      ),
    );
}
