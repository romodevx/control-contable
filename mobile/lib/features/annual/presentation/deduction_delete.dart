import 'dart:async';

import 'package:control_contable/api/generated/models/personal_deduction.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/records/stacked_undo.dart';
import 'package:control_contable/features/annual/application/annual_input_writes.dart';
import 'package:control_contable/features/annual/presentation/annual_write_feedback.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Deletes [deduction] at once, then offers Undo in the persistent toast
/// (deletes stack). Allowed to a read-only user, as records.
Future<void> deletePersonalDeduction(
  BuildContext context,
  WidgetRef ref,
  PersonalDeduction deduction,
) async {
  // The toast and its Undo outlive the list: keep what they need now.
  final messenger = ScaffoldMessenger.of(context);
  final l10n = AppLocalizations.of(context);
  final undo = ref.read(annualDeductionsUndoProvider);
  final writes = ref.read(annualInputWritesProvider);
  try {
    await writes.deleteDeduction(deduction);
  } on ApiFailure catch (failure) {
    undo.notify(
      messenger,
      refusalText(failure, l10n, fallback: l10n.annualDeductionsDeleteFailed),
    );
    return;
  }
  undo.push(
    deduction,
    messenger: messenger,
    message: l10n.annualDeductionsDeleted,
    undoLabel: l10n.recordsUndo,
    onUndo: (items) =>
        unawaited(_restore(messenger, l10n, undo, writes, items)),
  );
}

Future<void> _restore(
  ScaffoldMessengerState messenger,
  AppLocalizations l10n,
  StackedUndo<PersonalDeduction> undo,
  AnnualInputWrites writes,
  List<PersonalDeduction> deductions,
) async {
  final report = await writes.restoreDeductions(deductions);
  final refusal = report.refusals.firstOrNull;
  undo.notify(
    messenger,
    refusal == null
        ? l10n.annualDeductionsRestored(report.restored)
        : refusalText(
            refusal,
            l10n,
            fallback: l10n.annualDeductionsRestoreFailed,
            limitReached: l10n.annualDeductionsRestoreFull,
          ),
  );
}
