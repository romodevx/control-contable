import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flag_gate.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/annual/presentation/annual_entry_button.dart';
import 'package:control_contable/features/auth/presentation/account_button.dart';
import 'package:control_contable/features/auth/presentation/sign_out_button.dart';
import 'package:control_contable/features/dashboard/presentation/dashboard_home.dart';
import 'package:control_contable/features/expenses/presentation/expenses_entry_button.dart';
import 'package:control_contable/features/history/presentation/history_entry_button.dart';
import 'package:control_contable/features/incomes/presentation/incomes_entry_button.dart';
import 'package:control_contable/features/reminders/presentation/reminders_entry_button.dart';
import 'package:control_contable/features/tax_engine/presentation/open_estimate_button.dart';
import 'package:control_contable/features/taxpayers/presentation/add_taxpayer_button.dart';
import 'package:control_contable/features/taxpayers/presentation/no_regime_prompt.dart';
import 'package:control_contable/features/taxpayers/presentation/onboarding_launcher.dart';
import 'package:control_contable/features/taxpayers/presentation/taxpayer_switcher.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Home. With `module_dashboard` on it is the dashboard of the active
/// taxpayer (M6); off, the previous placeholder home stays. Offers
/// sign-out while `module_auth` is on (the router only shows home to a
/// signed-in user then). With `module_taxpayers` on, its title is the
/// active-taxpayer switcher and a user with no taxpayer is sent to the
/// onboarding. With `module_history` on, its app bar opens History; with
/// `module_annual` on, the annual estimate.
class HomeScreen extends StatelessWidget {
  /// Creates the screen.
  const new({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(
        title: ActiveTaxpayerSwitcher(fallback: Text(l10n.navHome)),
        actions: const [
          FlagGate(module: AppModule.annual, child: AnnualEntryButton()),
          FlagGate(module: AppModule.history, child: HistoryEntryButton()),
          FlagGate(module: AppModule.reminders, child: RemindersEntryButton()),
          FlagGate(module: AppModule.auth, child: AccountButton()),
          FlagGate(module: AppModule.auth, child: SignOutButton()),
        ],
      ),
      body: const FlagGate(
        module: AppModule.dashboard,
        fallback: _PlaceholderHome(),
        child: DashboardHome(),
      ),
    );
  }
}

/// The home before the dashboard: entry buttons into each module, and the
/// add-a-regime prompt when no regime is in effect.
class _PlaceholderHome extends StatelessWidget {
  const new();

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Padding(
      padding: const EdgeInsets.all(AppSpacing.gutter),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          // The placeholder talks about the estimate, so it is part of
          // the tax engine slice: hidden with it (MODULE-tax-engine §7).
          FlagGate(
            module: AppModule.taxEngine,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Text(
                  l10n.homePlaceholder,
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
                const SizedBox(height: AppSpacing.xl),
                const OpenEstimateButton(),
              ],
            ),
          ),
          const FlagGate(module: AppModule.taxpayers, child: NoRegimePrompt()),
          const FlagGate(
            module: AppModule.incomes,
            child: Padding(
              padding: EdgeInsets.only(bottom: AppSpacing.xs),
              child: IncomesEntryButton(),
            ),
          ),
          const FlagGate(
            module: AppModule.expenses,
            child: Padding(
              padding: EdgeInsets.only(bottom: AppSpacing.xs),
              child: ExpensesEntryButton(),
            ),
          ),
          const FlagGate(
            module: AppModule.taxpayers,
            child: AddTaxpayerButton(),
          ),
          const FlagGate(
            module: AppModule.taxpayers,
            child: OnboardingLauncher(),
          ),
        ],
      ),
    );
  }
}
