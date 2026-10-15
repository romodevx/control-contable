import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/reminders/application/reminder_settings.dart';
import 'package:control_contable/features/reminders/presentation/reminders_view.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Ajustes → Recordatorios (M9-10): the opt-in, the reminder details and
/// the next due dates. A page, not a dialog: it holds a list besides the
/// five settings (`ui-forms-and-modals`, field count and content).
class RemindersScreen extends ConsumerWidget {
  /// Creates the screen.
  const new({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final settings = ref.watch(reminderSettingsProvider);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.remindersTitle)),
      body: switch (settings) {
        AsyncData(:final value) => RemindersView(settings: value),
        AsyncError() => Padding(
          padding: const EdgeInsets.all(AppSpacing.gutter),
          child: LoadFailed(
            message: l10n.remindersLoadFailed,
            onRetry: () => ref.invalidate(reminderSettingsProvider),
          ),
        ),
        _ => Padding(
          padding: const EdgeInsets.all(AppSpacing.gutter),
          child: LoadStatus(message: l10n.remindersLoading, busy: true),
        ),
      },
    );
  }
}
