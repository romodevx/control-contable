import 'dart:async';

import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/reminders/presentation/system_settings_opener.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Notifications are blocked in the OS: what that means and the way to
/// allow them. The app never re-prompts (MODULE-reminders §7).
class PushDeniedNotice extends ConsumerWidget {
  /// Creates the notice.
  const new({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        FormNotice(message: l10n.remindersDenied),
        const SizedBox(height: AppSpacing.xs),
        OutlinedButton(
          onPressed: () => unawaited(ref.read(systemSettingsOpenerProvider)()),
          child: Text(l10n.remindersOpenSystemSettings),
        ),
      ],
    );
  }
}
