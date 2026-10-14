import 'package:control_contable/api/generated/models/regime_estimate_status.dart';
import 'package:control_contable/api/generated/models/tax_result.dart';
import 'package:control_contable/api/generated/models/tax_result_outcome.dart';
import 'package:control_contable/api/generated/models/taxpayer_estimate.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/tax_engine/domain/iva.dart';
import 'package:control_contable/features/tax_engine/presentation/money_row.dart';
import 'package:control_contable/features/tax_engine/presentation/status_chip.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// The month's totals on the navy hero: ISR, the one consolidated IVA (or
/// its saldo a favor), the withholdings to remit apart, the total to pay
/// and the due date. "Incompleto" when any regime misses inputs.
class EstimateTotalsCard extends StatelessWidget {
  /// The totals of [estimate].
  const new({required this.estimate, super.key});

  /// The estimate shown.
  final TaxpayerEstimate estimate;

  @override
  Widget build(BuildContext context) {
    final tokens = context.tokens;
    final colors = tokens.colors;
    final l10n = AppLocalizations.of(context);
    final totals = estimate.totals;
    final isIncomplete = estimate.regimes.any(
      (r) => r.status == RegimeEstimateStatus.incomplete,
    );
    final ivaContributors = estimate.regimes
        .where((r) => contributesIva(r.ivaComponents))
        .length;
    final muted = Theme.of(context).textTheme.bodyMedium
        ?.copyWith(color: colors.heroMuted);
    return DecoratedBox(
      decoration: BoxDecoration(
        color: colors.heroBg,
        borderRadius: BorderRadius.circular(tokens.radii.card),
      ),
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.md),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (isIncomplete) ...[
              Align(
                alignment: AlignmentDirectional.centerStart,
                child: StatusChip(
                  label: l10n.estimateIncomplete,
                  color: colors.heroText,
                ),
              ),
              const SizedBox(height: AppSpacing.xxs),
              Text(l10n.estimateTotalsIncomplete, style: muted),
              const SizedBox(height: AppSpacing.xs),
            ],
            _taxRow(
              totals.isr,
              l10n.estimateIsrPayable,
              l10n.estimateIsrInFavor,
            ),
            _taxRow(
              totals.iva,
              l10n.estimateIvaPayable,
              l10n.estimateIvaInFavor,
            ),
            if (ivaContributors > 1)
              Text(l10n.estimateIvaConsolidated(ivaContributors), style: muted),
            if (totals.withholdingsToRemitCents > 0)
              MoneyRow(
                label: l10n.estimateWithholdingsToRemit,
                centavos: totals.withholdingsToRemitCents,
                color: colors.heroText,
              ),
            Divider(color: colors.heroMuted, height: AppSpacing.xl),
            HeroAmount(
              label: l10n.estimateTotalPayable,
              centavos: totals.totalPayableRoundedCents,
              color: colors.heroText,
            ),
            const SizedBox(height: AppSpacing.xs),
            Text(
              l10n.estimateDueDate(estimate.dueDate),
              style: Theme.of(context).textTheme.bodyLarge
                  ?.copyWith(color: colors.heroText),
            ),
            Text(l10n.estimateDueDateNote, style: muted),
          ],
        ),
      ),
    );
  }

  /// A total as "a pagar" or, for a negative result, "saldo a favor"
  /// (never counted as a payment, PRD US-X1).
  Widget _taxRow(TaxResult result, String payable, String inFavor) => Builder(
    builder: (context) {
      final isInFavor = result.outcome == TaxResultOutcome.inFavor;
      return MoneyRow(
        label: isInFavor ? inFavor : payable,
        centavos: isInFavor
            ? result.amountCents.abs()
            : result.payableRoundedCents,
        color: context.tokens.colors.heroText,
      );
    },
  );
}
