import 'package:control_contable/api/generated/models/annual_balance_kind.dart';
import 'package:control_contable/api/generated/models/annual_estimate.dart';
import 'package:control_contable/api/generated/models/annual_section.dart';
import 'package:control_contable/api/generated/models/annual_section_status.dart';
import 'package:control_contable/api/generated/models/annual_tax.dart';
import 'package:control_contable/api/generated/models/annual_year_status.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/annual/domain/annual_headline.dart';
import 'package:control_contable/features/annual/presentation/annual_breakdown_button.dart';
import 'package:control_contable/features/annual/presentation/annual_labels.dart';
import 'package:control_contable/features/tax_engine/domain/money_format.dart';
import 'package:control_contable/features/tax_engine/presentation/money_row.dart';
import 'package:control_contable/features/tax_engine/presentation/status_chip.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// The headline on the navy hero: each headline section's balance in words
/// and figures (a PM's two sections each on its own, never netted), its
/// deadline and the way to its breakdown. Read as one sentence per balance
/// ("Declaración anual 2026, estimación. Saldo a favor, 9,408 pesos. Vence
/// el 30 de abril de 2027."). With no headline section, says so: no balance
/// is invented.
class AnnualHeadlineCard extends StatelessWidget {
  /// The headline of [estimate].
  const new({required this.estimate, super.key});

  /// The estimate shown.
  final AnnualEstimate estimate;

  @override
  Widget build(BuildContext context) {
    final tokens = context.tokens;
    final sections = headlineSections(estimate);
    return DecoratedBox(
      decoration: BoxDecoration(
        color: tokens.colors.heroBg,
        borderRadius: BorderRadius.circular(tokens.radii.card),
      ),
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.md),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          spacing: AppSpacing.md,
          children: [
            if (sections.isEmpty) _NoBalance(deadline: estimate.deadline),
            for (final (:section, :tax) in sections)
              _Balance(
                section: section,
                tax: tax,
                estimate: estimate,
                showsTitle: sections.length > 1,
              ),
          ],
        ),
      ),
    );
  }
}

class _Balance extends StatelessWidget {
  const new({
    required this.section,
    required this.tax,
    required this.estimate,
    required this.showsTitle,
  });

  final AnnualSection section;
  final AnnualTax tax;
  final AnnualEstimate estimate;
  final bool showsTitle;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final colors = context.tokens.colors;
    final text = Theme.of(context).textTheme;
    final balance = tax.balance;
    final kind = balanceKindText(l10n, balance.kind);
    final deadline = section.deadline ?? estimate.deadline;
    final dueOn = deadline == null ? null : parseBusinessDate(deadline);
    final isIncomplete = section.status == AnnualSectionStatus.incomplete;
    final inProgress = estimate.yearStatus == AnnualYearStatus.inProgress;
    final summary = [
      l10n.annualSpokenHeading('${estimate.fiscalYear}'),
      if (balance.kind == AnnualBalanceKind.aCargo ||
          balance.kind == AnnualBalanceKind.aFavor)
        l10n.annualSpokenBalance(kind, spokenMoney(balance.amountCents))
      else
        '$kind.',
      if (dueOn != null) l10n.annualSpokenDeadline(dueOn),
      if (isIncomplete) l10n.annualSpokenIncomplete,
      if (inProgress) l10n.annualSpokenInProgress,
    ].join(' ');
    final muted = text.bodyMedium?.copyWith(color: colors.heroMuted);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Semantics(
          container: true,
          label: summary,
          excludeSemantics: true,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            spacing: AppSpacing.xxs,
            children: [
              if (showsTitle)
                Text(
                  sectionTitle(l10n, section),
                  style: text.titleSmall?.copyWith(color: colors.heroText),
                ),
              if (isIncomplete || inProgress)
                Wrap(
                  spacing: AppSpacing.xs,
                  runSpacing: AppSpacing.xxs,
                  children: [
                    if (isIncomplete)
                      StatusChip(
                        label: l10n.estimateIncomplete,
                        color: colors.heroText,
                      ),
                    if (inProgress)
                      StatusChip(
                        label: l10n.annualYearInProgress,
                        color: colors.heroText,
                      ),
                  ],
                ),
              if (isIncomplete)
                Text(l10n.estimateTotalsIncomplete, style: muted),
              HeroAmount(
                label: kind,
                centavos: balance.amountCents,
                color: colors.heroText,
              ),
              if (dueOn != null) ...[
                Text(
                  l10n.estimateDueDate(dueOn),
                  style: text.bodyLarge?.copyWith(color: colors.heroText),
                ),
                Text(l10n.estimateDueDateNote, style: muted),
              ],
            ],
          ),
        ),
        AnnualBreakdownButton(
          section: section.kind,
          taxpayerId: estimate.taxpayerId,
          fiscalYear: estimate.fiscalYear,
          color: colors.heroText,
        ),
      ],
    );
  }
}

class _NoBalance extends StatelessWidget {
  const new({required this.deadline});

  final String? deadline;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final colors = context.tokens.colors;
    final style = Theme.of(context).textTheme.bodyLarge
        ?.copyWith(color: colors.heroText);
    final dueOn = deadline;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      spacing: AppSpacing.xxs,
      children: [
        Text(l10n.annualNoHeadline, style: style),
        if (dueOn != null)
          Text(l10n.estimateDueDate(parseBusinessDate(dueOn)), style: style),
      ],
    );
  }
}
