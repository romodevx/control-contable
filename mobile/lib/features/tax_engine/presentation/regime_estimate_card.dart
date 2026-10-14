import 'dart:async';

import 'package:control_contable/api/generated/models/cpa_topic.dart';
import 'package:control_contable/api/generated/models/regime_estimate.dart';
import 'package:control_contable/api/generated/models/regime_estimate_status.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/tax_engine/domain/estimate_periods.dart';
import 'package:control_contable/features/tax_engine/presentation/missing_inputs_block.dart';
import 'package:control_contable/features/tax_engine/presentation/regime_amounts.dart';
import 'package:control_contable/features/tax_engine/presentation/status_chip.dart';
import 'package:control_contable/features/tax_engine/tax_engine_routes.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// One regime of the month: name, status chip, its figures (marked when
/// they depend on missing inputs), the explanation of what it does not
/// estimate, CPA-unverified topics, and links to its breakdown and inputs.
/// An informational regime shows only its explanation: no amount, no
/// breakdown (PRD US-X1).
class RegimeEstimateCard extends StatelessWidget {
  /// The card of [regime] in [period] of [taxpayerId].
  const new({
    required this.regime,
    required this.taxpayerId,
    required this.period,
    super.key,
  });

  /// The regime estimate.
  final RegimeEstimate regime;

  /// Whose estimate it is.
  final String taxpayerId;

  /// The month (`YYYY-MM`) of the estimate.
  final String period;

  bool get _isInformational =>
      regime.status == RegimeEstimateStatus.informational ||
      regime.status == RegimeEstimateStatus.$unknown;

  @override
  Widget build(BuildContext context) {
    final tokens = context.tokens;
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    final isIncomplete = regime.status == RegimeEstimateStatus.incomplete;
    final explanation = regime.explanation;
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
            Semantics(
              header: true,
              child: Text(regime.regimeName, style: text.titleMedium),
            ),
            Row(
              children: [
                Expanded(
                  child: Text(
                    l10n.estimateRegimeCode(regime.regimeCode),
                    style: text.bodyMedium?.copyWith(
                      color: tokens.colors.muted,
                    ),
                  ),
                ),
                StatusChip.of(context, regime.status),
              ],
            ),
            if (regime.cpaReviewPending)
              Text(
                l10n.estimateCpaTopics(_topics(l10n, regime.cpaTopics)),
                style: text.bodyMedium,
              ),
            if (explanation != null) Text(explanation, style: text.bodyMedium),
            if (isIncomplete)
              Text(l10n.estimateIncompleteFigures, style: text.titleSmall),
            if (!_isInformational) RegimeAmounts(regime: regime),
            if (isIncomplete && regime.missingInputs.isNotEmpty)
              MissingInputsBlock(
                missing: regime.missingInputs,
                location: TaxEngineRoutes.taxInputs(
                  taxpayerId,
                  regime.regimeCode,
                  fiscalYear: fiscalYearOf(period),
                  field: regime.missingInputs.first.key,
                ),
              ),
            if (!_isInformational)
              Align(
                alignment: AlignmentDirectional.centerEnd,
                child: TextButton(
                  onPressed: () => unawaited(
                    context.push<void>(
                      TaxEngineRoutes.breakdown(
                        taxpayerId,
                        period,
                        regime.regimeCode,
                      ),
                    ),
                  ),
                  child: Text(l10n.estimateViewBreakdown),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

String _topics(AppLocalizations l10n, List<CpaTopic> topics) => topics
    .map(
      (topic) => switch (topic) {
        CpaTopic.rifReduction => l10n.cpaTopicRifReduction,
        CpaTopic.rifBimonthlyIva => l10n.cpaTopicRifBimonthlyIva,
        CpaTopic.platformDefinitiveIva => l10n.cpaTopicPlatformDefinitiveIva,
        CpaTopic.crossRegimeIvaNetting => l10n.cpaTopicCrossRegimeIvaNetting,
        CpaTopic.resicoPmInterestWithholding =>
          l10n.cpaTopicResicoPmInterestWithholding,
        CpaTopic.$unknown => l10n.cpaTopicOther,
      },
    )
    .toSet()
    .join(', ');
