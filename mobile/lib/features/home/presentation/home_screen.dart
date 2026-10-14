import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flag_gate.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/auth/presentation/account_button.dart';
import 'package:control_contable/features/auth/presentation/sign_out_button.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Placeholder home until the dashboard module (M6) ships. Offers sign-out
/// while `module_auth` is on (the router only shows home to a signed-in
/// user then).
class HomeScreen extends StatelessWidget {
  /// Creates the screen.
  const new({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.navHome),
        actions: const [
          FlagGate(module: AppModule.auth, child: AccountButton()),
          FlagGate(module: AppModule.auth, child: SignOutButton()),
        ],
      ),
      body: Padding(
        padding: const EdgeInsets.all(AppSpacing.gutter),
        child: Text(
          l10n.homePlaceholder,
          style: Theme.of(context).textTheme.bodyMedium,
        ),
      ),
    );
  }
}
