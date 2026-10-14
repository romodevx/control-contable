import 'dart:async';

import 'package:control_contable/api/generated/models/disclaimer_severity.dart';
import 'package:control_contable/api/generated/models/taxpayer_estimate.dart';
import 'package:control_contable/api/generated/models/taxpayer_estimate_status.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/estimate_legend.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_notices.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_totals_card.dart';
import 'package:control_contable/features/tax_engine/presentation/status_chip.dart';
import 'package:control_contable/features/tax_engine/tax_engine_routes.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// This month's estimate on Home (MODULE-dashboard US-D1), in reading
/// order: the strong CPA banner, the totals (the M5 card, reused), the
/// mandatory legend right under them, then the taxpayer status and the
/// way to the estimate screen. Only the strong disclaimer travels here;
/// the rest are one tap away.
class DashboardHero extends StatelessWidget {
  /// The hero of [estimate] with the taxpayer-level [status].
  const new({
    required this.estimate,
    required this.status,
    required this.legend,
    super.key,
  });

  /// The server's estimate for the month.
  final TaxpayerEstimate estimate;

  /// The taxpayer-level status (the trend's last point).
  final TaxpayerEstimateStatus status;

  /// The server's legend.
  final String legend;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final informational = status == TaxpayerEstimateStatus.informational;
    final strong = estimate.disclaimers.where(
      (d) => d.severity == DisclaimerSeverity.strong,
    );
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      spacing: AppSpacing.sm,
      children: [
        for (final disclaimer in strong)
          CpaReviewBanner(
            key: const ValueKey('cpa-review-banner'),
            disclaimer: disclaimer,
          ),
        if (informational)
          Text(
            l10n.dashboardInformational,
            style: Theme.of(context).textTheme.bodyLarge,
          )
        else ...[
          Semantics(
            container: true,
            child: GestureDetector(
              key: const ValueKey('dashboard-hero'),
              excludeFromSemantics: true,
              onTap: () => _openEstimate(context),
              child: EstimateTotalsCard(estimate: estimate),
            ),
          ),
          EstimateLegend(text: legend),
        ],
        _StatusLine(
          key: const ValueKey('dashboard-status'),
          status: status,
          onOpen: () => _openEstimate(context),
        ),
      ],
    );
  }

  /// The estimate screen opens on the current month, the one Home shows.
  static void _openEstimate(BuildContext context) =>
      unawaited(context.push<void>(TaxEngineRoutes.estimate));
}

/// The status chip and the link to the estimate screen: "Completa tus
/// datos" when inputs are missing (the totals card already says
/// "Incompleto"), else "Ver detalle".
class _StatusLine extends StatelessWidget {
  const new({required this.status, required this.onOpen, super.key});

  final TaxpayerEstimateStatus status;
  final VoidCallback onOpen;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final chip = switch (status) {
      TaxpayerEstimateStatus.complete => l10n.estimateComplete,
      TaxpayerEstimateStatus.partial => l10n.estimatePartial,
      TaxpayerEstimateStatus.informational => l10n.estimateInformational,
      _ => null,
    };
    final isIncomplete = status == TaxpayerEstimateStatus.incomplete;
    return Wrap(
      spacing: AppSpacing.sm,
      runSpacing: AppSpacing.xs,
      crossAxisAlignment: WrapCrossAlignment.center,
      children: [
        if (chip != null) StatusChip(label: chip),
        TextButton(
          onPressed: onOpen,
          child: Text(
            isIncomplete ? l10n.dashboardCompleteData : l10n.dashboardSeeDetail,
          ),
        ),
      ],
    );
  }
}
