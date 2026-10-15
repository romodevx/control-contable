import 'package:control_contable/api/generated/models/counterparty_kind.dart';
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/core/fiscal/withholding_labels.dart';
import 'package:control_contable/core/fiscal/withholding_preview_loader.dart';
import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/widgets/derived_withholding_card.dart';
import 'package:control_contable/features/incomes/application/income_form_model.dart';
import 'package:control_contable/features/incomes/domain/income_preview_query.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// What the income card says (MODULE-withholdings §2.3): who withholds
/// what and what reaches the user, from the server's derivation; the
/// warnings the facts already imply show at once, even offline.
WithholdingCardCopy incomeWithholdingCopy(
  IncomeFormModel model,
  WithholdingPreviewLoader<IncomePreviewQuery> loader,
  AppLocalizations l10n,
) {
  final preview = loader.preview;
  final local = [
    if (model.facts.kind == CounterpartyKind.foreign)
      l10n.withholdingWarningForeign,
    if (model.platformWithoutRfc) l10n.withholdingWarningNoRfc,
  ];
  final warnings = {
    ...local,
    for (final warning in preview?.warnings ?? const <Warning>[])
      withholdingWarningText(warning, l10n),
  }.toList();
  final notice = switch (loader.status) {
    WithholdingPreviewStatus.incomplete => l10n.incomeWithholdingIncomplete,
    WithholdingPreviewStatus.failed when preview == null =>
      l10n.withholdingOffline,
    WithholdingPreviewStatus.failed => l10n.withholdingStale,
    WithholdingPreviewStatus.current => null,
  };
  if (preview == null) {
    return WithholdingCardCopy(warnings: warnings, notice: notice);
  }
  final withholding = preview.withholding;
  final withheld = withholding.isrCents + withholding.ivaCents > 0;
  return WithholdingCardCopy(
    headline: [
      if (withheld)
        l10n.incomeWithheldLine(
          formatPesos(withholding.isrCents),
          formatPesos(withholding.ivaCents),
        )
      else
        l10n.incomeNothingWithheld,
      l10n.incomeReceiveLine(formatPesos(withholding.netCents)),
    ],
    explanations: [for (final line in withholding.explanations) line.text],
    warnings: warnings,
    notice: notice,
  );
}

/// The read-only card of the income form, following [loader].
class IncomeWithholdingCard extends StatelessWidget {
  /// The card of [model].
  const new({required this.model, required this.loader, super.key});

  /// The form whose facts it derives from.
  final IncomeFormModel model;

  /// The debounced server derivation.
  final WithholdingPreviewLoader<IncomePreviewQuery> loader;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return ListenableBuilder(
      listenable: loader,
      builder: (context, _) => DerivedWithholdingCard(
        copy: incomeWithholdingCopy(model, loader, l10n),
        helpLabel: l10n.withholdingHelpLabel,
        helpText: l10n.withholdingHelpText,
      ),
    );
  }
}
