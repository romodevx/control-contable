import 'dart:async';

import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/taxpayers/taxpayers_routes.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// Heights of the skeleton blocks: banner, hero, month card, chart.
const List<double> _skeletonBlocks = [64, 220, 140, 240];

/// Home while it loads: an announced status and grey blocks (not
/// announced). Never the previous taxpayer's figures.
class DashboardSkeleton extends StatelessWidget {
  /// Creates the skeleton.
  const new({super.key});

  @override
  Widget build(BuildContext context) {
    final tokens = context.tokens;
    return Column(
      key: const ValueKey('dashboard-skeleton'),
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        LoadStatus(
          message: AppLocalizations.of(context).dashboardLoading,
          busy: true,
        ),
        for (final height in _skeletonBlocks)
          Padding(
            padding: const EdgeInsets.only(top: AppSpacing.md),
            child: ExcludeSemantics(
              child: DecoratedBox(
                decoration: BoxDecoration(
                  color: tokens.colors.border,
                  borderRadius: BorderRadius.circular(tokens.radii.card),
                ),
                child: SizedBox(height: height),
              ),
            ),
          ),
      ],
    );
  }
}

/// A failed Home load: why, and *Reintentar*. A partial Home is never
/// shown (MODULE-dashboard §5).
class DashboardFailure extends StatelessWidget {
  /// The failure of [error], retried by [onRetry].
  const new({required this.error, required this.onRetry, super.key});

  /// What the load threw.
  final Object error;

  /// Loads again.
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return LoadFailed(
      message: error is ApiUnreachable
          ? l10n.estimateUnreachable
          : l10n.dashboardLoadFailed,
      onRetry: onRetry,
    );
  }
}

/// No records yet (MODULE-dashboard §2.7): a light message instead of a
/// big $0.00 hero; the quick actions follow it.
class DashboardEmpty extends StatelessWidget {
  /// Creates the empty state.
  const new({super.key});

  @override
  Widget build(BuildContext context) => Text(
    AppLocalizations.of(context).dashboardEmpty,
    key: const ValueKey('dashboard-empty'),
    style: Theme.of(context).textTheme.bodyLarge,
  );
}

/// No active taxpayer (every one archived): the way to register one.
class DashboardNoTaxpayer extends StatelessWidget {
  /// Creates the prompt.
  const new({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text(
          l10n.dashboardNoTaxpayer,
          style: Theme.of(context).textTheme.bodyLarge,
        ),
        const SizedBox(height: AppSpacing.md),
        FilledButton(
          onPressed: () =>
              unawaited(context.push<void>(TaxpayersRoutes.onboarding)),
          child: Text(l10n.dashboardCompleteOnboarding),
        ),
      ],
    );
  }
}
