import 'package:control_contable/api/generated/models/annual_breakdown_line.dart';
import 'package:control_contable/api/generated/models/annual_breakdown_line_operation.dart';
import 'package:control_contable/api/generated/models/annual_computed_section_kind.dart';
import 'package:control_contable/api/generated/models/annual_section_kind.dart';
import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/annual/application/annual_estimate.dart';
import 'package:control_contable/features/annual/presentation/annual_input_card.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// How the breakdown names a cap that binds (MODULE-annual §2.7).
const String _capPrefix = 'Tope aplicado';

/// Registered vs applied: everything registered in the year, what the
/// estimate's `pf_art152` section applied after the legal caps and cash
/// exclusions, and every "Tope aplicado" line of its breakdown. What the
/// estimate cannot answer is left out; the registered total always shows.
class DeductionsSummary extends ConsumerWidget {
  /// The summary of [query] with [registeredCents] registered.
  const new({required this.query, required this.registeredCents, super.key});

  /// Whose deductions and which year.
  final AnnualQuery query;

  /// Σ registered amounts of the year.
  final int registeredCents;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    final estimate = ref.watch(annualEstimateProvider(query)).value;
    final section = estimate?.sections
        .where((s) => s.kind == AnnualSectionKind.pfArt152)
        .firstOrNull;
    final applied = section?.tax?.personalDeductionsCents;
    final caps = section == null ? const <String>[] : _caps(ref);
    return AnnualInputCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          AnnualAmountRow(
            label: l10n.annualDeductionsRegisteredTotal,
            amount: formatPesos(registeredCents),
          ),
          if (applied != null) ...[
            const SizedBox(height: AppSpacing.sm),
            AnnualAmountRow(
              label: l10n.annualDeductionsAppliedTotal,
              amount: formatPesos(applied),
            ),
            Text(
              l10n.annualDeductionsAppliedHelp,
              style: text.bodySmall?.copyWith(
                color: context.tokens.colors.muted,
              ),
            ),
          ],
          if (caps.isNotEmpty) ...[
            const SizedBox(height: AppSpacing.md),
            Semantics(
              header: true,
              child: Text(
                l10n.annualDeductionsCapsTitle,
                style: text.titleSmall,
              ),
            ),
            for (final cap in caps) Text(cap, style: text.bodyMedium),
          ],
        ],
      ),
    );
  }

  List<String> _caps(WidgetRef ref) {
    final breakdown = ref
        .watch(
          annualBreakdownProvider((
            taxpayerId: query.taxpayerId,
            fiscalYear: query.fiscalYear,
            section: AnnualComputedSectionKind.pfArt152,
          )),
        )
        .value;
    return [
      for (final line in breakdown?.lines ?? const <AnnualBreakdownLine>[])
        if (line.operation == AnnualBreakdownLineOperation.info &&
            line.concept.startsWith(_capPrefix))
          line.concept,
    ];
  }
}
