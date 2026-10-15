import 'package:control_contable/api/generated/models/annual_tax.dart';
import 'package:control_contable/api/generated/models/payment_source_counts.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/annual/presentation/annual_labels.dart';
import 'package:control_contable/features/tax_engine/presentation/money_row.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// A computed section's annual ISR: income, personal deductions (PF), the
/// base (PF) or resultado fiscal (PM), the ISR, each credit, where the
/// provisional payments come from, and the balance in words. Figures are
/// marked when the section is incomplete.
class AnnualTaxFigures extends StatelessWidget {
  /// The figures of [tax]; [isPersonaMoral] names the base.
  const new({
    required this.tax,
    required this.isPersonaMoral,
    required this.isIncomplete,
    super.key,
  });

  /// The section's annual ISR.
  final AnnualTax tax;

  /// Whether the base is a PM's resultado fiscal.
  final bool isPersonaMoral;

  /// Whether the figures depend on missing inputs.
  final bool isIncomplete;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    final credits = tax.credits;
    final deductions = tax.personalDeductionsCents;
    final sources = _sourcesText(l10n, credits.paymentSources);
    Widget row(String label, int centavos) =>
        MoneyRow(label: label, centavos: centavos, isIncomplete: isIncomplete);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        row(l10n.annualAccumulableIncome, tax.accumulableIncomeCents),
        if (deductions != null) row(l10n.annualPersonalDeductions, deductions),
        row(
          isPersonaMoral ? l10n.annualPmBase : l10n.annualBase,
          tax.baseCents,
        ),
        row(l10n.annualIsr, tax.isrCents),
        row(l10n.annualPayments, credits.paymentsCents),
        if (sources != null)
          Text(
            sources,
            style: text.bodyMedium?.copyWith(
              color: context.tokens.colors.muted,
            ),
          ),
        row(l10n.annualWithholdings, credits.withholdingsCents),
        if (credits.salaryIsrCents > 0)
          row(l10n.annualSalaryIsr, credits.salaryIsrCents),
        const Divider(),
        MoneyRow(
          label: balanceKindText(l10n, tax.balance.kind),
          centavos: tax.balance.amountCents,
          labelStyle: text.titleSmall,
          isIncomplete: isIncomplete,
        ),
      ],
    );
  }

  /// "Pagos: 6 meses confirmados, 3 estimados por la app."; `null` when no
  /// month was counted.
  static String? _sourcesText(
    AppLocalizations l10n,
    PaymentSourceCounts sources,
  ) {
    final parts = [
      if (sources.filedMonths > 0)
        l10n.annualPaymentsFiled(sources.filedMonths),
      if (sources.taxInputMonths > 0)
        l10n.annualPaymentsTaxInput(sources.taxInputMonths),
      if (sources.estimatedByAppMonths > 0)
        l10n.annualPaymentsByApp(sources.estimatedByAppMonths),
    ];
    return parts.isEmpty ? null : l10n.annualPaymentsLine(parts.join(', '));
  }
}
