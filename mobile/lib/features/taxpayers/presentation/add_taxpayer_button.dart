import 'dart:async';

import 'package:control_contable/features/taxpayers/taxpayers_routes.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// Starts the onboarding of a new taxpayer. Gate it with
/// `FlagGate(module: AppModule.taxpayers, …)`.
class AddTaxpayerButton extends StatelessWidget {
  /// Creates the button.
  const new({super.key});

  @override
  Widget build(BuildContext context) => FilledButton.icon(
    icon: const Icon(Icons.person_add_alt_outlined),
    label: Text(AppLocalizations.of(context).homeAddTaxpayer),
    onPressed: () => unawaited(context.push<void>(TaxpayersRoutes.onboarding)),
  );
}
