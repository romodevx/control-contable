import 'dart:async';

import 'package:control_contable/features/annual/annual_routes.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// Opens the annual estimate from the app bar. Gate it with
/// `FlagGate(module: AppModule.annual, …)` (`useFlag`).
class AnnualEntryButton extends StatelessWidget {
  /// Creates the button.
  const new({super.key});

  @override
  Widget build(BuildContext context) => IconButton(
    tooltip: AppLocalizations.of(context).annualTitle,
    icon: const Icon(Icons.event_note_outlined),
    onPressed: () => unawaited(context.push<void>(AnnualRoutes.estimate)),
  );
}
