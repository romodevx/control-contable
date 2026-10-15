import 'package:control_contable/api/generated/models/history_year_totals.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/tax_engine/presentation/money_row.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// "Suma de estimados mensuales": the year's ISR, IVA payable,
/// withholdings to remit (when any) and total, as the server summed them.
/// A sum of provisional estimates, never the annual return (§10).
class HistoryYearTotalsCard extends StatelessWidget {
  /// The card of [totals].
  const new({required this.totals, super.key});

  /// The server's sums.
  final HistoryYearTotals totals;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final tokens = context.tokens;
    final text = Theme.of(context).textTheme;
    return DecoratedBox(
      key: const ValueKey('history-year-totals'),
      decoration: BoxDecoration(
        color: tokens.colors.surface,
        border: Border.all(color: tokens.colors.border),
        borderRadius: BorderRadius.circular(tokens.radii.card),
      ),
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.md),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Semantics(
              header: true,
              child: Text(l10n.historyYearTotalsTitle, style: text.titleMedium),
            ),
            const SizedBox(height: AppSpacing.xs),
            MoneyRow(
              label: l10n.estimateIsrPayable,
              centavos: totals.isrPayableRoundedCents,
            ),
            MoneyRow(
              label: l10n.estimateIvaPayable,
              centavos: totals.ivaPayableRoundedCents,
            ),
            if (totals.withholdingsToRemitCents > 0)
              MoneyRow(
                label: l10n.estimateWithholdingsToRemit,
                centavos: totals.withholdingsToRemitCents,
              ),
            Divider(color: tokens.colors.border),
            MoneyRow(
              label: l10n.estimateTotalPayable,
              centavos: totals.totalPayableRoundedCents,
            ),
            const SizedBox(height: AppSpacing.xxs),
            Text(
              l10n.historyYearTotalsNote,
              style: text.bodyMedium?.copyWith(color: tokens.colors.muted),
            ),
          ],
        ),
      ),
    );
  }
}
