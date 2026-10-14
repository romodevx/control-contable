import 'package:control_contable/api/generated/models/iva_components.dart';
import 'package:control_contable/api/generated/models/regime_estimate.dart';
import 'package:control_contable/api/generated/models/tax_result_outcome.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/tax_engine/domain/iva.dart';
import 'package:control_contable/features/tax_engine/presentation/money_row.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// A regime's own figures: its ISR (or saldo a favor), its contribution to
/// the month's single IVA (never an amount to pay by itself, §2.6) and the
/// withholdings it must remit.
class RegimeAmounts extends StatelessWidget {
  /// The figures of [regime].
  const new({required this.regime, super.key});

  /// The regime estimate.
  final RegimeEstimate regime;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final isr = regime.isr;
    final iva = regime.ivaComponents;
    final isrInFavor = isr?.outcome == TaxResultOutcome.inFavor;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (isr != null)
          MoneyRow(
            label: isrInFavor
                ? l10n.estimateIsrInFavor
                : l10n.estimateIsrPayable,
            centavos: isrInFavor
                ? isr.amountCents.abs()
                : isr.payableRoundedCents,
            color: isrInFavor ? context.tokens.colors.favor : null,
          ),
        if (iva != null && contributesIva(iva))
          _IvaComponentRows(components: iva),
        if (regime.withheldToRemitCents > 0)
          MoneyRow(
            label: l10n.estimateWithholdingsToRemit,
            centavos: regime.withheldToRemitCents,
          ),
      ],
    );
  }
}

class _IvaComponentRows extends StatelessWidget {
  const new({required this.components});

  final IvaComponents components;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    final muted = context.tokens.colors.muted;
    final rows = <(String, int)>[
      (l10n.estimateIvaCharged, components.chargedCollectedCents),
      (l10n.estimateIvaExcluded, -components.definitiveExcludedCents),
      (l10n.estimateIvaCreditable, -components.creditablePaidCents),
      (l10n.estimateIvaWithheld, -components.withheldFromTaxpayerCents),
    ];
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        for (final (label, cents) in rows)
          if (cents != 0)
            MoneyRow(
              label: label,
              centavos: cents,
              labelStyle: text.bodyMedium,
              amountStyle: text.bodyMedium,
            ),
        Text(
          l10n.estimateIvaContributes,
          style: text.bodyMedium?.copyWith(color: muted),
        ),
      ],
    );
  }
}
