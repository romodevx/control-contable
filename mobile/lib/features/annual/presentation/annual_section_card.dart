import 'package:control_contable/api/generated/models/annual_section.dart';
import 'package:control_contable/api/generated/models/annual_section_status.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/annual/presentation/annual_breakdown_button.dart';
import 'package:control_contable/features/annual/presentation/annual_labels.dart';
import 'package:control_contable/features/annual/presentation/annual_missing_inputs.dart';
import 'package:control_contable/features/annual/presentation/annual_section_extras.dart';
import 'package:control_contable/features/annual/presentation/annual_tax_figures.dart';
import 'package:control_contable/features/tax_engine/presentation/regime_estimate_card.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// One section of the annual estimate: its title, regimes and status, the
/// CPA topics, whether it must be filed and why, its own deadline, then by
/// kind its ISR figures, the RESICO informative check, the PM informative
/// figures or the explanation; its notes, what is missing (each with the
/// way to complete it) and the way to its breakdown.
class AnnualSectionCard extends StatelessWidget {
  /// The card of [section] in [taxpayerId]'s [fiscalYear].
  const new({
    required this.section,
    required this.taxpayerId,
    required this.fiscalYear,
    required this.isPersonaMoral,
    required this.showsBreakdownLink,
    super.key,
  });

  /// The section.
  final AnnualSection section;

  /// Whose estimate it is.
  final String taxpayerId;

  /// The fiscal year estimated.
  final int fiscalYear;

  /// Whether the taxpayer is a persona moral.
  final bool isPersonaMoral;

  /// Whether the card links to its breakdown (the headline already does
  /// for its own sections).
  final bool showsBreakdownLink;

  @override
  Widget build(BuildContext context) {
    final tokens = context.tokens;
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    final isIncomplete = section.status == AnnualSectionStatus.incomplete;
    final obligation = obligationText(l10n, section.obligation);
    final explanation = section.explanation;
    final deadline = section.deadline;
    final tax = section.tax;
    final resico = section.resicoCheck;
    final extras = section.pmExtras;
    return DecoratedBox(
      decoration: BoxDecoration(
        color: tokens.colors.surface,
        border: Border.all(color: tokens.colors.border),
        borderRadius: BorderRadius.circular(tokens.radii.card),
      ),
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.md),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          spacing: AppSpacing.xs,
          children: [
            _SectionHeader(section: section),
            if (section.cpaReviewPending)
              Text(
                l10n.estimateCpaTopics(cpaTopicsText(l10n, section.cpaTopics)),
                style: text.bodyMedium,
              ),
            if (obligation != null) Text(obligation, style: text.titleSmall),
            for (final reason in section.obligationReasons)
              Text(reason, style: text.bodyMedium),
            if (explanation != null) Text(explanation, style: text.bodyMedium),
            if (deadline != null)
              Text(
                l10n.annualDeadline(parseBusinessDate(deadline)),
                style: text.bodyMedium,
              ),
            if (isIncomplete)
              Text(l10n.estimateIncompleteFigures, style: text.titleSmall),
            if (tax != null)
              AnnualTaxFigures(
                tax: tax,
                isPersonaMoral: isPersonaMoral,
                isIncomplete: isIncomplete,
              ),
            if (resico != null) ResicoCheckFigures(check: resico),
            if (extras != null)
              PmExtrasBlock(
                key: const ValueKey('pm-extras'),
                extras: extras,
                fiscalYear: fiscalYear,
                isUnverified: section.cpaReviewPending,
              ),
            for (final note in section.notes)
              Text(note.text, style: text.bodyMedium),
            if (section.missingInputs.isNotEmpty)
              AnnualMissingInputs(
                missing: section.missingInputs,
                taxpayerId: taxpayerId,
                fiscalYear: fiscalYear,
              ),
            if (showsBreakdownLink)
              AnnualBreakdownButton(
                section: section.kind,
                taxpayerId: taxpayerId,
                fiscalYear: fiscalYear,
              ),
          ],
        ),
      ),
    );
  }
}

/// The section's title, the regimes it covers and its status chip.
class _SectionHeader extends StatelessWidget {
  const new({required this.section});

  final AnnualSection section;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      spacing: AppSpacing.xxs,
      children: [
        Semantics(
          header: true,
          child: Text(sectionTitle(l10n, section), style: text.titleMedium),
        ),
        Wrap(
          spacing: AppSpacing.xs,
          runSpacing: AppSpacing.xxs,
          crossAxisAlignment: WrapCrossAlignment.center,
          children: [
            Text(
              l10n.annualSectionRegimes(section.regimes.join(', ')),
              style: text.bodyMedium?.copyWith(
                color: context.tokens.colors.muted,
              ),
            ),
            annualStatusChip(context, section.status),
          ],
        ),
      ],
    );
  }
}
