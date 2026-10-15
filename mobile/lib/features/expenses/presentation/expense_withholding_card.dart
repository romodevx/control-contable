import 'package:control_contable/api/generated/models/supplier_kind.dart';
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/core/fiscal/withholding_labels.dart';
import 'package:control_contable/core/fiscal/withholding_preview_loader.dart';
import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/core/widgets/derived_withholding_card.dart';
import 'package:control_contable/features/expenses/application/expense_form_model.dart';
import 'package:control_contable/features/expenses/domain/expense_preview_query.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// What the expense card says (MODULE-withholdings §2.3): what a persona
/// moral must withhold, what the supplier receives, what goes to the SAT
/// and by when, and the deduction note, from the server's derivation.
WithholdingCardCopy expenseWithholdingCopy(
  ExpenseFormModel model,
  WithholdingPreviewLoader<ExpensePreviewQuery> loader,
  AppLocalizations l10n,
) {
  final preview = loader.preview;
  final withholding = preview?.withholding;
  final remitDueOn = withholding?.remitDueOn;
  final warnings = {
    if (model.supplier.kind == SupplierKind.foreign)
      l10n.withholdingWarningForeign,
    for (final warning in preview?.warnings ?? const <Warning>[])
      withholdingWarningText(warning, l10n, remitDueOn: remitDueOn),
  }.toList();
  final notice = switch (loader.status) {
    WithholdingPreviewStatus.incomplete => l10n.expenseWithholdingIncomplete,
    WithholdingPreviewStatus.failed when preview == null =>
      l10n.withholdingOffline,
    WithholdingPreviewStatus.failed => l10n.withholdingStale,
    WithholdingPreviewStatus.current => null,
  };
  if (withholding == null) {
    return WithholdingCardCopy(warnings: warnings, notice: notice);
  }
  final withheldCents = withholding.isrCents + withholding.ivaCents;
  return WithholdingCardCopy(
    headline: [
      if (withheldCents > 0)
        l10n.expenseMustWithholdLine(
          formatPesos(withholding.isrCents),
          formatPesos(withholding.ivaCents),
        )
      else
        l10n.expenseNothingToWithhold,
      l10n.expensePaySupplierLine(formatPesos(withholding.netCents)),
    ],
    details: [
      if (withheldCents > 0 && remitDueOn != null)
        l10n.expenseRemitLine(
          formatPesos(withheldCents),
          parseBusinessDate(remitDueOn),
        ),
    ],
    explanations: [for (final line in withholding.explanations) line.text],
    warnings: warnings,
    notice: notice,
  );
}

/// The read-only card of a persona moral's expense, following [loader].
class ExpenseWithholdingCard extends StatelessWidget {
  /// The card of [model].
  const new({required this.model, required this.loader, super.key});

  /// The form whose facts it derives from.
  final ExpenseFormModel model;

  /// The debounced server derivation.
  final WithholdingPreviewLoader<ExpensePreviewQuery> loader;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return ListenableBuilder(
      listenable: loader,
      builder: (context, _) => DerivedWithholdingCard(
        copy: expenseWithholdingCopy(model, loader, l10n),
        helpLabel: l10n.withholdingHelpLabel,
        helpText: l10n.withholdingHelpText,
      ),
    );
  }
}
