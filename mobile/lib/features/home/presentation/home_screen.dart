import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flag_gate.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/auth/auth_routes.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// Placeholder home until the dashboard module (M6) ships. Offers sign-up
/// while `module_auth` is on.
class HomeScreen extends StatelessWidget {
  /// Creates the screen.
  const new({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.navHome)),
      body: Padding(
        padding: const EdgeInsets.all(AppSpacing.gutter),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              l10n.homePlaceholder,
              style: Theme.of(context).textTheme.bodyMedium,
            ),
            FlagGate(
              module: AppModule.auth,
              child: Padding(
                padding: const EdgeInsets.only(top: AppSpacing.xl),
                child: OutlinedButton(
                  onPressed: () => context.push<void>(AuthRoutes.register),
                  child: Text(l10n.homeCreateAccount),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
