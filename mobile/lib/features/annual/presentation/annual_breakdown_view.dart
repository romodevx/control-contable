import 'package:control_contable/api/generated/models/annual_breakdown.dart';
import 'package:control_contable/api/generated/models/annual_section_status.dart';
import 'package:control_contable/api/generated/models/disclaimer.dart';
import 'package:control_contable/api/generated/models/disclaimer_code.dart';
import 'package:control_contable/api/generated/models/disclaimer_severity.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/core/widgets/estimate_legend.dart';
import 'package:control_contable/features/annual/presentation/annual_labels.dart';
import 'package:control_contable/features/annual/presentation/annual_line_tile.dart';
import 'package:control_contable/features/annual/presentation/annual_missing_inputs.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_notices.dart';
import 'package:control_contable/features/tax_engine/presentation/regime_estimate_card.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// A loaded annual breakdown, in the monthly breakdown's reading order:
/// the header, the strong CPA warning above any figure, the mandatory
/// legend, the warnings, what is missing (each with the way to complete
/// it), the lines in `seq` order, then the informational disclaimers.
class AnnualBreakdownView extends StatelessWidget {
  /// Shows [breakdown].
  const new({required this.breakdown, super.key});

  /// The breakdown shown.
  final AnnualBreakdown breakdown;

  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.stretch,
    spacing: AppSpacing.md,
    children: [
      for (final block in _blocks(AppLocalizations.of(context), context))
        Semantics(container: true, child: block),
    ],
  );

  List<Widget> _blocks(AppLocalizations l10n, BuildContext context) {
    final section = breakdown.section;
    final text = Theme.of(context).textTheme;
    final isIncomplete = section.status == AnnualSectionStatus.incomplete;
    // The server sends them so; sorting (O(n log n)) guards the order.
    final lines = [...breakdown.lines]..sort((a, b) => a.seq.compareTo(b.seq));
    return [
      _Header(breakdown: breakdown),
      for (final disclaimer in _bySeverity(DisclaimerSeverity.strong))
        CpaReviewBanner(key: _keyOf(disclaimer), disclaimer: disclaimer),
      if (section.cpaReviewPending)
        Text(
          l10n.estimateCpaTopics(cpaTopicsText(l10n, section.cpaTopics)),
          style: text.bodyMedium,
        ),
      EstimateLegend(text: breakdown.legend),
      for (final disclaimer in _bySeverity(DisclaimerSeverity.warning))
        DisclaimerWarning(key: _keyOf(disclaimer), disclaimer: disclaimer),
      if (section.missingInputs.isNotEmpty)
        AnnualMissingInputs(
          missing: section.missingInputs,
          taxpayerId: breakdown.taxpayerId,
          fiscalYear: breakdown.fiscalYear,
        ),
      if (lines.isEmpty)
        Text(l10n.annualBreakdownNoLines, style: text.bodyLarge)
      else
        Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            for (final line in lines)
              AnnualLineTile(
                key: ValueKey('line-${line.seq}'),
                line: line,
                dependsOnMissingInputs: isIncomplete,
              ),
          ],
        ),
      for (final disclaimer in breakdown.disclaimers.where(_isNote))
        DisclaimerNote(key: _keyOf(disclaimer), disclaimer: disclaimer),
    ];
  }

  static bool _isNote(Disclaimer disclaimer) =>
      disclaimer.severity == DisclaimerSeverity.info ||
      disclaimer.severity == DisclaimerSeverity.$unknown;

  static Key _keyOf(Disclaimer disclaimer) =>
      disclaimer.code == DisclaimerCode.cpaReviewPending
      ? const ValueKey('cpa-review-banner')
      : ValueKey('disclaimer-${disclaimer.code.json ?? disclaimer.text}');

  Iterable<Disclaimer> _bySeverity(DisclaimerSeverity severity) =>
      breakdown.disclaimers.where((d) => d.severity == severity);
}

/// Whose computation this is: the section, its year and regimes, its
/// status, and the parameter set and engine version used.
class _Header extends StatelessWidget {
  const new({required this.breakdown});

  final AnnualBreakdown breakdown;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    final section = breakdown.section;
    final parameters = section.parameterSet;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      spacing: AppSpacing.xxs,
      children: [
        Semantics(
          header: true,
          child: Text(sectionTitle(l10n, section), style: text.titleLarge),
        ),
        Wrap(
          spacing: AppSpacing.xs,
          runSpacing: AppSpacing.xxs,
          crossAxisAlignment: WrapCrossAlignment.center,
          children: [
            Text(
              l10n.annualBreakdownSubtitle(
                '${breakdown.fiscalYear}',
                section.regimes.join(', '),
              ),
              style: text.bodyMedium,
            ),
            annualStatusChip(context, section.status),
          ],
        ),
        Text(
          l10n.breakdownVersions(
            '${parameters.fiscalYear}',
            '${parameters.version}',
            breakdown.engineVersion,
          ),
          style: text.bodySmall?.copyWith(color: context.tokens.colors.muted),
        ),
      ],
    );
  }
}
