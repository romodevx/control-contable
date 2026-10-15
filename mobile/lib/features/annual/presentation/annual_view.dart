import 'package:control_contable/api/generated/models/annual_estimate.dart';
import 'package:control_contable/api/generated/models/annual_section_kind.dart';
import 'package:control_contable/api/generated/models/disclaimer.dart';
import 'package:control_contable/api/generated/models/disclaimer_code.dart';
import 'package:control_contable/api/generated/models/disclaimer_severity.dart';
import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/estimate_legend.dart';
import 'package:control_contable/features/annual/presentation/annual_headline_card.dart';
import 'package:control_contable/features/annual/presentation/annual_inputs_block.dart';
import 'package:control_contable/features/annual/presentation/annual_section_card.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_notices.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_states.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// A loaded annual estimate, in the reading order of MODULE-annual §7: the
/// strong CPA banner above every figure, the mandatory legend right above
/// the headline balance, the warnings, "Datos para tu anual", one card per
/// section, then the informational disclaimers.
class AnnualView extends StatelessWidget {
  /// Shows [estimate].
  const new({required this.estimate, super.key});

  /// The estimate shown.
  final AnnualEstimate estimate;

  @override
  Widget build(BuildContext context) {
    if (estimate.sections.isEmpty) {
      return EstimateMessage(
        message: AppLocalizations.of(context)
            .annualNoSections('${estimate.fiscalYear}'),
      );
    }
    // Each block is its own screen-reader stop (see EstimateView).
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      spacing: AppSpacing.md,
      children: [
        for (final block in _blocks()) Semantics(container: true, child: block),
      ],
    );
  }

  List<Widget> _blocks() {
    final headline = {for (final kind in estimate.headlineSections) kind.json};
    return [
      for (final disclaimer in _bySeverity(DisclaimerSeverity.strong))
        CpaReviewBanner(key: _keyOf(disclaimer), disclaimer: disclaimer),
      // The legend sits right on the figures it qualifies.
      Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        spacing: AppSpacing.xs,
        children: [
          EstimateLegend(text: estimate.legend),
          AnnualHeadlineCard(
            key: const ValueKey('annual-headline'),
            estimate: estimate,
          ),
        ],
      ),
      for (final disclaimer in _bySeverity(DisclaimerSeverity.warning))
        DisclaimerWarning(key: _keyOf(disclaimer), disclaimer: disclaimer),
      AnnualInputsBlock(
        key: const ValueKey('annual-inputs'),
        estimate: estimate,
      ),
      for (final section in estimate.sections)
        AnnualSectionCard(
          key: ValueKey(
            section.kind == AnnualSectionKind.explanation
                ? 'annual-section-explanation-${section.regimes.join('-')}'
                : 'annual-section-${section.kind.json}',
          ),
          section: section,
          taxpayerId: estimate.taxpayerId,
          fiscalYear: estimate.fiscalYear,
          isPersonaMoral: estimate.personType == PersonType.moral,
          showsBreakdownLink: !headline.contains(section.kind.json),
        ),
      for (final disclaimer in estimate.disclaimers.where(_isNote))
        DisclaimerNote(key: _keyOf(disclaimer), disclaimer: disclaimer),
    ];
  }

  static bool _isNote(Disclaimer disclaimer) =>
      disclaimer.severity == DisclaimerSeverity.info ||
      disclaimer.severity == DisclaimerSeverity.$unknown;

  /// Codes are unique per estimate; an unknown one keys by its text.
  static Key _keyOf(Disclaimer disclaimer) =>
      disclaimer.code == DisclaimerCode.cpaReviewPending
      ? const ValueKey('cpa-review-banner')
      : ValueKey('disclaimer-${disclaimer.code.json ?? disclaimer.text}');

  Iterable<Disclaimer> _bySeverity(DisclaimerSeverity severity) =>
      estimate.disclaimers.where((d) => d.severity == severity);
}
