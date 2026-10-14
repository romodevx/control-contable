import 'dart:async';

import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/core/access/paywall_sheet.dart';
import 'package:control_contable/core/records/record_save.dart';
import 'package:control_contable/core/records/record_save_toast.dart';
import 'package:control_contable/core/records/stacked_undo.dart';
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
String incomeSaveNoticeText(RecordSaveNotice notice, AppLocalizations l10n) =>
    switch (notice) {
      RecordSaveNotice.unreachable => l10n.incomeSaveUnreachable,
      RecordSaveNotice.idReused => l10n.incomeSaveIdReused,
      RecordSaveNotice.gone => l10n.incomeSaveGone,
      RecordSaveNotice.failed => l10n.incomeSaveFailed,
    };

/// Confirms a save ([updated] for an edit) on [messenger] (see
/// [showRecordSavedToast]): warnings never block, and a pending Undo of
/// [undo] is never closed by it.
void showIncomeSavedToast(
  ScaffoldMessengerState messenger,
  AppLocalizations l10n,
  List<Warning> warnings, {
  required StackedUndo<Income> undo,
  bool updated = false,
  bool overForm = false,
}) {
  final count = warnings.length;
  showRecordSavedToast(
    messenger,
    heading: switch ((count > 0, updated)) {
      (true, false) => l10n.incomeSavedWithWarnings(count),
      (true, true) => l10n.incomeUpdatedWithWarnings(count),
      (false, false) => l10n.incomeSaved,
      (false, true) => l10n.incomeUpdated,
    },
    notes: [for (final w in warnings) describeIncomeWarning(w, l10n)],
    undo: undo,
    overForm: overForm,
  );
}
