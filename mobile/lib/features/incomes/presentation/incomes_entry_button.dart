import 'dart:async';

import 'package:control_contable/features/incomes/incomes_routes.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// Opens the income list. Gate it with
/// `FlagGate(module: AppModule.incomes, …)`.
class IncomesEntryButton extends StatelessWidget {
  /// Creates the button.
  const new({super.key});

  @override
  Widget build(BuildContext context) => OutlinedButton(
    child: Text(AppLocalizations.of(context).navIncome),
    onPressed: () => unawaited(context.push<void>(IncomesRoutes.list)),
  );
}
