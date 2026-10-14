import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/auth/auth_routes.dart';
import 'package:control_contable/features/auth/presentation/session_wipe_notice.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// Shown once the account is deleted: says it is done, reminds that a
/// store subscription is cancelled in the store, and leads to login.
class GoodbyeScreen extends StatelessWidget {
  /// Creates the screen.
  const new({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.goodbyeTitle),
        automaticallyImplyLeading: false,
      ),
      body: SafeArea(
        child: ListView(
          padding: const EdgeInsets.all(AppSpacing.gutter),
          children: [
            const SessionWipeNotice(
              padding: EdgeInsets.only(bottom: AppSpacing.md),
            ),
            Text(l10n.goodbyeBody, style: text.bodyLarge),
            const SizedBox(height: AppSpacing.md),
            Text(l10n.accountDeleteSubscription, style: text.bodyMedium),
            const SizedBox(height: AppSpacing.xl),
            FilledButton(
              onPressed: () => context.go(AuthRoutes.login),
              child: Text(l10n.goodbyeToLogin),
            ),
          ],
        ),
      ),
    );
  }
}
