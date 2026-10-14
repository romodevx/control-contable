import 'dart:async';

import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/tax_engine/tax_engine_routes.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// Home's way into the monthly estimate until the dashboard (M6) shows it
/// inline. Place it behind `FlagGate(module: AppModule.taxEngine, …)`.
class OpenEstimateButton extends StatelessWidget {
  /// Creates the button.
  const new({super.key});

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: AppSpacing.lg),
    child: FilledButton(
      onPressed: () => unawaited(context.push<void>(TaxEngineRoutes.estimate)),
      child: Text(AppLocalizations.of(context).homeOpenEstimate),
    ),
  );
}
