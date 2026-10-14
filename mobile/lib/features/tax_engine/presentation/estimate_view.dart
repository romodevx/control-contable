import 'package:control_contable/api/generated/models/disclaimer.dart';
import 'package:control_contable/api/generated/models/disclaimer_code.dart';
import 'package:control_contable/api/generated/models/disclaimer_severity.dart';
import 'package:control_contable/api/generated/models/taxpayer_estimate.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/estimate_legend.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_notices.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_states.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_totals_card.dart';
import 'package:control_contable/features/tax_engine/presentation/regime_estimate_card.dart';
import 'package:control_contable/features/taxpayers/domain/periods.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// A loaded estimate, in reading order: the strong CPA warning above the
/// figures, the totals, the mandatory legend right under them, the
/// warnings, one card per regime, then the informational disclaimers
/// (MODULE-tax-engine §8).
class EstimateView extends StatelessWidget {
  /// Shows [estimate].
  const new({required this.estimate, super.key});

  /// The estimate shown.
  final TaxpayerEstimate estimate;

  @override
  Widget build(BuildContext context) {
    if (estimate.regimes.isEmpty) {
      final month = MaterialLocalizations.of(context)
          .formatMonthYear(periodMonth(estimate.period));
      return EstimateMessage(
        message: AppLocalizations.of(context).estimateNoRegimes(month),
      );
    }
    final strong = _bySeverity(DisclaimerSeverity.strong);
    final warnings = _bySeverity(DisclaimerSeverity.warning);
    final notes = estimate.disclaimers.where(
      (d) =>
          d.severity == DisclaimerSeverity.info ||
          d.severity == DisclaimerSeverity.$unknown,
    );
    // Each block is its own screen-reader stop; without containers the
    // whole estimate merges into one node read as a single blob.
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      spacing: AppSpacing.md,
      children: [
        for (final block in _blocks(strong, warnings, notes))
          Semantics(container: true, child: block),
      ],
    );
  }

  List<Widget> _blocks(
    Iterable<Disclaimer> strong,
    Iterable<Disclaimer> warnings,
    Iterable<Disclaimer> notes,
  ) => [
    for (final disclaimer in strong)
      CpaReviewBanner(key: _keyOf(disclaimer), disclaimer: disclaimer),
    EstimateTotalsCard(
      key: const ValueKey('estimate-totals'),
      estimate: estimate,
    ),
    EstimateLegend(text: estimate.legend),
    for (final disclaimer in warnings)
      DisclaimerWarning(key: _keyOf(disclaimer), disclaimer: disclaimer),
    for (final regime in estimate.regimes)
      RegimeEstimateCard(
        key: ValueKey('regime-${regime.regimeCode}'),
        regime: regime,
        taxpayerId: estimate.taxpayerId,
        period: estimate.period,
      ),
    for (final disclaimer in notes)
      DisclaimerNote(key: _keyOf(disclaimer), disclaimer: disclaimer),
  ];

  /// Codes are unique per estimate; an unknown one keys by its text.
  static Key _keyOf(Disclaimer disclaimer) =>
      disclaimer.code == DisclaimerCode.cpaReviewPending
      ? const ValueKey('cpa-review-banner')
      : ValueKey('disclaimer-${disclaimer.code.json ?? disclaimer.text}');

  Iterable<Disclaimer> _bySeverity(DisclaimerSeverity severity) =>
      estimate.disclaimers.where((d) => d.severity == severity);
}
