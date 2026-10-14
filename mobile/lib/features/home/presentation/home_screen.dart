import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flag_gate.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/auth/presentation/account_button.dart';
import 'package:control_contable/features/auth/presentation/sign_out_button.dart';
import 'package:control_contable/features/tax_engine/presentation/open_estimate_button.dart';
import 'package:control_contable/features/taxpayers/presentation/add_taxpayer_button.dart';
import 'package:control_contable/features/taxpayers/presentation/no_regime_prompt.dart';
import 'package:control_contable/features/taxpayers/presentation/onboarding_launcher.dart';
import 'package:control_contable/features/taxpayers/presentation/taxpayer_switcher.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Placeholder home until the dashboard module (M6) ships. Offers sign-out
/// while `module_auth` is on (the router only shows home to a signed-in
/// user then). With `module_taxpayers` on, its title is the active-taxpayer
/// switcher, a user with no taxpayer is sent to the onboarding, and one
/// whose active taxpayer has no regime in effect is offered to add one.
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
          FlagGate(module: AppModule.auth, child: AccountButton()),
          FlagGate(module: AppModule.auth, child: SignOutButton()),
        ],
      ),
      body: Padding(
        padding: const EdgeInsets.all(AppSpacing.gutter),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              l10n.homePlaceholder,
              style: Theme.of(context).textTheme.bodyMedium,
            ),
            const SizedBox(height: AppSpacing.xl),
            const FlagGate(
              module: AppModule.taxEngine,
              child: OpenEstimateButton(),
            ),
            const FlagGate(
              module: AppModule.taxpayers,
              child: NoRegimePrompt(),
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
      ),
    );
  }
}
