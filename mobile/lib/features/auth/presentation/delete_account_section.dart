import 'dart:async';

import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/auth/presentation/delete_account_dialog.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// *Eliminar cuenta* (in-app deletion, a store requirement): a heading, one
/// line on what it does, and the button that opens the confirmation.
class DeleteAccountSection extends StatelessWidget {
  /// Creates the section.
  const new({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final danger = theme.colorScheme.error;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Semantics(
          header: true,
          child: Text(
            l10n.accountDeleteHeading,
            style: theme.textTheme.titleMedium,
          ),
        ),
        const SizedBox(height: AppSpacing.xs),
        Text(l10n.accountDeleteIntro, style: theme.textTheme.bodyMedium),
        const SizedBox(height: AppSpacing.sm),
        OutlinedButton(
          style: OutlinedButton.styleFrom(
            foregroundColor: danger,
            side: BorderSide(color: danger),
          ),
          onPressed: () => unawaited(showDeleteAccountDialog(context)),
          child: Text(l10n.accountDeleteOpen),
        ),
      ],
    );
  }
}
