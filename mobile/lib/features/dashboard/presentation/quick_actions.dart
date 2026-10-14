import 'dart:async';

import 'package:control_contable/core/access/paywall_sheet.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flag_gate.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/expenses/expenses_routes.dart';
import 'package:control_contable/features/incomes/incomes_routes.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// "Registrar ingreso" (primary) and "Registrar gasto" (secondary), each
/// behind its module's flag. They open the M3/M4 create pages for the
/// active taxpayer and run [onReturn] when the user comes back, so Home
/// shows the new figures. A read-only user gets the paywall instead (UX
/// only: the server's `402` is the boundary).
class QuickActions extends StatelessWidget {
  /// The actions; [readOnly] opens the paywall instead.
  const new({required this.readOnly, required this.onReturn, super.key});

  /// Whether the user can only read.
  final bool readOnly;

  /// Home reloads.
  final VoidCallback onReturn;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        FlagGate(
          module: AppModule.incomes,
          child: Padding(
            padding: const EdgeInsets.only(bottom: AppSpacing.xs),
            child: FilledButton(
              onPressed: () => _open(context, IncomesRoutes.create),
              child: Text(l10n.dashboardRegisterIncome),
            ),
          ),
        ),
        FlagGate(
          module: AppModule.expenses,
          child: OutlinedButton(
            onPressed: () => _open(context, ExpensesRoutes.create),
            child: Text(l10n.dashboardRegisterExpense),
          ),
        ),
      ],
    );
  }

  void _open(BuildContext context, String location) {
    if (readOnly) {
      final l10n = AppLocalizations.of(context);
      unawaited(
        showPaywallSheet(
          context,
          title: l10n.onboardingPaywallTitle,
          body: l10n.paywallWriteBody,
        ),
      );
      return;
    }
    unawaited(context.push<void>(location).whenComplete(onReturn));
  }
}
