import 'package:control_contable/api/generated/models/pm_annual_extras.dart';
import 'package:control_contable/api/generated/models/resico_pf_check.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/tax_engine/domain/rate_percent.dart';
import 'package:control_contable/features/tax_engine/presentation/computation_line_tile.dart';
import 'package:control_contable/features/tax_engine/presentation/money_row.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// A PM's informative figures, apart from the balance: the PTU base, the
/// PTU to distribute and the suggested coeficiente de utilidad for next
/// year, marked ⚠️ while a CPA has not validated them.
class PmExtrasBlock extends StatelessWidget {
  /// The block of [extras] for [fiscalYear].
  const new({
    required this.extras,
    required this.fiscalYear,
    required this.isUnverified,
    super.key,
  });

  /// The informative figures.
  final PmAnnualExtras extras;

  /// The fiscal year estimated (the coeficiente is for the next one).
  final int fiscalYear;

  /// Whether a rule behind these figures awaits a CPA.
  final bool isUnverified;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    final toDistribute = extras.ptuToDistributeCents;
    final cu = extras.suggestedCuNextYear;
    return _InformativeBox(
      children: [
        Semantics(
          header: true,
          child: Text(l10n.annualPmExtrasTitle, style: text.titleSmall),
        ),
        if (isUnverified) const CpaReviewMark(),
        MoneyRow(label: l10n.annualPtuBase, centavos: extras.ptuBaseCents),
        if (toDistribute != null)
          MoneyRow(label: l10n.annualPtuToDistribute, centavos: toDistribute),
        // A coeficiente is a plain ratio, shown as the server sends it.
        if (cu != null)
          Text(
            l10n.annualSuggestedCu('${fiscalYear + 1}', cu),
            style: text.bodyLarge,
          ),
      ],
    );
  }
}

/// RESICO PF's 113-F check: income × the annual rate − the monthly
/// payments − the 1.25 % withheld, and the difference, said to be neither
/// payable nor refundable (rule 3.13.7).
class ResicoCheckFigures extends StatelessWidget {
  /// The figures of [check].
  const new({required this.check, super.key});

  /// The informative check.
  final ResicoPfCheck check;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    return _InformativeBox(
      children: [
        MoneyRow(label: l10n.annualResicoIncome, centavos: check.incomeCents),
        Text(
          l10n.annualResicoRate(formatRatePercent(check.rate)),
          style: text.bodyMedium,
        ),
        MoneyRow(label: l10n.annualResicoIsr, centavos: check.annualIsrCents),
        MoneyRow(
          label: l10n.annualResicoMonthly,
          centavos: check.monthlyPaymentsCents,
        ),
        MoneyRow(
          label: l10n.annualResicoWithheld,
          centavos: check.withheldCents,
        ),
        MoneyRow(
          label: l10n.annualResicoDifference,
          centavos: check.differenceCents,
          labelStyle: text.titleSmall,
        ),
        Text(l10n.annualResicoDifferenceHelp, style: text.bodyMedium),
      ],
    );
  }
}

/// Informative figures on the legend's tint: never styled as a balance.
class _InformativeBox extends StatelessWidget {
  const new({required this.children});

  final List<Widget> children;

  @override
  Widget build(BuildContext context) {
    final tokens = context.tokens;
    return DecoratedBox(
      decoration: BoxDecoration(
        color: tokens.colors.legendBg,
        borderRadius: BorderRadius.circular(tokens.radii.control),
      ),
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.sm),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          spacing: AppSpacing.xxs,
          children: children,
        ),
      ),
    );
  }
}
