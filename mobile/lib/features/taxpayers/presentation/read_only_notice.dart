import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Size of the lock icon.
const double _iconSize = 40;

/// What a read-only user (`402 entitlement_required`: trial over, no plan)
/// sees instead of the create: why, what still works, and the way out.
/// Nothing was created; no form is offered again.
class ReadOnlyNotice extends StatelessWidget {
  /// The notice; [onLeave] closes the onboarding.
  const new({required this.onLeave, super.key});

  /// Leaves the onboarding.
  final VoidCallback onLeave;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    return SingleChildScrollView(
      padding: const EdgeInsets.all(AppSpacing.gutter),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Align(
            alignment: Alignment.centerLeft,
            child: Icon(
              Icons.lock_outline,
              size: _iconSize,
              color: context.tokens.colors.primary,
            ),
          ),
          const SizedBox(height: AppSpacing.md),
          Semantics(
            header: true,
            liveRegion: true,
            child: Text(
              l10n.onboardingPaywallTitle,
              style: theme.textTheme.headlineMedium,
            ),
          ),
          const SizedBox(height: AppSpacing.xs),
          Text(l10n.onboardingPaywallBody, style: theme.textTheme.bodyMedium),
          const SizedBox(height: AppSpacing.xl),
          FilledButton(
            onPressed: onLeave,
            child: Text(l10n.onboardingPaywallAction),
          ),
        ],
      ),
    );
  }
}
