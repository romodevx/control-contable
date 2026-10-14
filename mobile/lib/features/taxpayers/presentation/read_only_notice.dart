import 'package:control_contable/features/taxpayers/presentation/stop_notice.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// What a read-only user (`402 entitlement_required`: trial over, no plan)
/// sees instead of the create: why, what still works, and the way out.
/// Nothing was created; no form is offered again.
class ReadOnlyNotice extends StatelessWidget {
  /// The notice; [onLeave] closes the flow. [body] says what needs a plan
  /// (registering a taxpayer by default).
  const new({required this.onLeave, this.body, this.actionLabel, super.key});

  /// Leaves the flow.
  final VoidCallback onLeave;

  /// What still works and what needs a plan.
  final String? body;

  /// The way out (*Volver al inicio* by default).
  final String? actionLabel;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return StopNotice(
      icon: Icons.lock_outline,
      title: l10n.onboardingPaywallTitle,
      body: body ?? l10n.onboardingPaywallBody,
      actionLabel: actionLabel ?? l10n.onboardingPaywallAction,
      onLeave: onLeave,
    );
  }
}
