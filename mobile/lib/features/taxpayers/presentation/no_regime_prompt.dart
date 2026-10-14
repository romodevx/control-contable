import 'dart:async';

import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/taxpayers/application/active_taxpayer.dart';
import 'package:control_contable/features/taxpayers/domain/periods.dart';
import 'package:control_contable/features/taxpayers/domain/regime_in_effect.dart';
import 'package:control_contable/features/taxpayers/taxpayers_routes.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// On home, when the active taxpayer has no regime in effect this month
/// (every one ended): says there is nothing to estimate and links to adding
/// a regime. Home stays; the onboarding is not reopened (it is for a user
/// with no taxpayer at all).
///
/// Renders nothing otherwise. Place it behind
/// `FlagGate(module: AppModule.taxpayers, …)`.
class NoRegimePrompt extends ConsumerWidget {
  /// Creates the prompt.
  const new({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final active = ref.watch(activeTaxpayerProvider).value;
    if (active == null) return const SizedBox.shrink();
    final period = currentPeriod(ref.watch(clockProvider)());
    if (hasRegimeInEffect(active.regimes, period: period)) {
      return const SizedBox.shrink();
    }
    final l10n = AppLocalizations.of(context);
    return Padding(
      padding: const EdgeInsets.only(bottom: AppSpacing.lg),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          // Static content of home, not a reaction to an action: read in
          // order, never announced (a live region would speak on every visit).
          Text(l10n.homeNoRegime, style: Theme.of(context).textTheme.bodyLarge),
          const SizedBox(height: AppSpacing.xs),
          OutlinedButton(
            onPressed: () => unawaited(
              context.push<void>(TaxpayersRoutes.addRegime(active.id)),
            ),
            child: Text(l10n.settingsAddRegime),
          ),
        ],
      ),
    );
  }
}
