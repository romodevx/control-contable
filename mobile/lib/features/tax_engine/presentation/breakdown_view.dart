import 'package:control_contable/api/generated/models/breakdown.dart';
import 'package:control_contable/api/generated/models/breakdown_line_tax.dart';
import 'package:control_contable/api/generated/models/cpa_topic.dart';
import 'package:control_contable/api/generated/models/disclaimer.dart';
import 'package:control_contable/api/generated/models/disclaimer_code.dart';
import 'package:control_contable/api/generated/models/disclaimer_severity.dart';
import 'package:control_contable/api/generated/models/regime_estimate_status.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/estimate_legend.dart';
import 'package:control_contable/features/tax_engine/domain/breakdown_groups.dart';
import 'package:control_contable/features/tax_engine/domain/estimate_periods.dart';
import 'package:control_contable/features/tax_engine/presentation/breakdown_group_section.dart';
import 'package:control_contable/features/tax_engine/presentation/breakdown_header.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_notices.dart';
import 'package:control_contable/features/tax_engine/presentation/missing_inputs_block.dart';
import 'package:control_contable/features/tax_engine/presentation/regime_estimate_card.dart';
import 'package:control_contable/features/tax_engine/tax_engine_routes.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// A loaded breakdown, in reading order: the header, the strong CPA
/// warning above any figure, the mandatory legend, the warnings, what is
/// missing (with the jump to the inputs form), the lines by tax, then the
/// informational disclaimers (MODULE-tax-engine §8).
class BreakdownView extends StatelessWidget {
  /// Shows [breakdown].
  const new({required this.breakdown, super.key});

  /// The breakdown shown.
  final Breakdown breakdown;

  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.stretch,
    spacing: AppSpacing.md,
    children: [
      // Each block is its own screen-reader stop (see EstimateView).
      for (final block in _blocks(AppLocalizations.of(context), context))
        Semantics(container: true, child: block),
    ],
  );

  List<Widget> _blocks(AppLocalizations l10n, BuildContext context) {
    final regime = breakdown.regime;
    final text = Theme.of(context).textTheme;
    final isIncomplete = regime.status == RegimeEstimateStatus.incomplete;
    final explanation = regime.explanation;
    final groups = groupLines(breakdown.lines);
    return [
      BreakdownHeader(breakdown: breakdown),
      for (final disclaimer in _bySeverity(DisclaimerSeverity.strong))
        CpaReviewBanner(key: _keyOf(disclaimer), disclaimer: disclaimer),
      if (regime.cpaReviewPending)
        Text(
          l10n.estimateCpaTopics(cpaTopicsText(l10n, regime.cpaTopics)),
          style: text.bodyMedium,
        ),
      EstimateLegend(text: breakdown.legend),
      for (final disclaimer in _bySeverity(DisclaimerSeverity.warning))
        DisclaimerWarning(key: _keyOf(disclaimer), disclaimer: disclaimer),
      if (explanation != null) Text(explanation, style: text.bodyLarge),
      if (isIncomplete && regime.missingInputs.isNotEmpty)
        MissingInputsBlock(
          missing: regime.missingInputs,
          location: TaxEngineRoutes.taxInputs(
            breakdown.taxpayerId,
            regime.regimeCode,
            fiscalYear: fiscalYearOf(breakdown.period),
            field: regime.missingInputs.first.key,
          ),
        ),
      if (groups.isEmpty) Text(l10n.breakdownNoLines, style: text.bodyLarge),
      for (final group in groups)
        BreakdownGroupSection(
          key: ValueKey('group-${group.tax.json}'),
          group: group,
          isIncomplete: isIncomplete,
          ivaNettingUnverified: _ivaNettingUnverified(group),
        ),
      for (final disclaimer in breakdown.disclaimers.where(_isNote))
        DisclaimerNote(key: _keyOf(disclaimer), disclaimer: disclaimer),
    ];
  }

  /// The month's IVA was netted across regimes (OQ-T2): the regime names
  /// the topic, or a consolidated IVA line needs CPA review for a reason
  /// other than the RIF's bimonthly IVA.
  bool _ivaNettingUnverified(BreakdownGroup group) {
    if (group.tax != BreakdownLineTax.iva) return false;
    final topics = breakdown.regime.cpaTopics;
    if (topics.contains(CpaTopic.crossRegimeIvaNetting)) return true;
    return !topics.contains(CpaTopic.rifBimonthlyIva) &&
        group.lines.any((line) => line.needsCpaReview);
  }

  static bool _isNote(Disclaimer disclaimer) =>
      disclaimer.severity == DisclaimerSeverity.info ||
      disclaimer.severity == DisclaimerSeverity.$unknown;

  /// Codes are unique per breakdown; an unknown one keys by its text.
  static Key _keyOf(Disclaimer disclaimer) =>
      disclaimer.code == DisclaimerCode.cpaReviewPending
      ? const ValueKey('cpa-review-banner')
      : ValueKey('disclaimer-${disclaimer.code.json ?? disclaimer.text}');

  Iterable<Disclaimer> _bySeverity(DisclaimerSeverity severity) =>
      breakdown.disclaimers.where((d) => d.severity == severity);
}
