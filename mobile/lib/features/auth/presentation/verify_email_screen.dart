import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/auth/application/pending_sign_up.dart';
import 'package:control_contable/features/auth/auth_routes.dart';
import 'package:control_contable/features/auth/presentation/code_step_form.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// Email verification (US-A1b, US-A1c, D8): the 6-digit code, sent with the
/// password and name chosen at sign-up; a resend after the cooldown. Every
/// code rejection reads the same (no account enumeration).
class VerifyEmailScreen extends ConsumerWidget {
  /// Creates the screen.
  const new({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final email = ref.watch(
      pendingSignUpProvider.select((draft) => draft?.email),
    );
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.authVerifyTitle)),
      body: email == null
          ? const _NoSignUp()
          : CodeStepForm(key: ValueKey(email), email: email),
    );
  }
}

class _NoSignUp extends StatelessWidget {
  const new();

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Padding(
      padding: const EdgeInsets.all(AppSpacing.gutter),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(
            l10n.authVerifyNoPending,
            style: Theme.of(context).textTheme.bodyLarge,
          ),
          const SizedBox(height: AppSpacing.lg),
          FilledButton(
            onPressed: () => context.go(AuthRoutes.register),
            child: Text(l10n.authRegisterTitle),
          ),
        ],
      ),
    );
  }
}
