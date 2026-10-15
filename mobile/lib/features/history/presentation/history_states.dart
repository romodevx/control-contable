import 'dart:async';

import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flag_gate.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/incomes/incomes_routes.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// Heights of the skeleton blocks: year totals, then three months.
const List<double> _skeletonBlocks = [160, 96, 96, 96];

/// History while it loads: an announced status and grey blocks (not
/// announced). Never another taxpayer's or year's figures.
class HistorySkeleton extends StatelessWidget {
  /// Creates the skeleton.
  const new({super.key});

  @override
  Widget build(BuildContext context) {
    final tokens = context.tokens;
    return Column(
      key: const ValueKey('history-skeleton'),
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        LoadStatus(
          message: AppLocalizations.of(context).historyLoading,
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

/// A failed History load: why, and *Reintentar*. No partial year is ever
/// shown (MODULE-history §5).
class HistoryFailure extends StatelessWidget {
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
          : l10n.historyLoadFailed,
      onRetry: onRetry,
    );
  }
}

/// No record in any month of the year (MODULE-history §7): the way to
/// start, when incomes are on.
class HistoryEmpty extends StatelessWidget {
  /// Creates the empty state.
  const new({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Column(
      key: const ValueKey('history-empty'),
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        HistoryMessage(message: l10n.historyEmpty),
        const SizedBox(height: AppSpacing.md),
        FlagGate(
          module: AppModule.incomes,
          child: FilledButton(
            onPressed: () =>
                unawaited(context.push<void>(IncomesRoutes.create)),
            child: Text(l10n.historyRegisterIncome),
          ),
        ),
      ],
    );
  }
}

/// A plain message where the year would be.
class HistoryMessage extends StatelessWidget {
  /// Reads [message].
  const new({required this.message, super.key});

  /// What the user reads.
  final String message;

  @override
  Widget build(BuildContext context) =>
      Text(message, style: Theme.of(context).textTheme.bodyLarge);
}
