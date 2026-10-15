import 'package:control_contable/api/generated/models/reminder_settings.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/reminders/application/push_messaging.dart';
import 'package:control_contable/features/reminders/application/reminder_settings.dart';
import 'package:control_contable/features/reminders/domain/push_token.dart';
import 'package:control_contable/features/reminders/presentation/push_denied_notice.dart';
import 'package:control_contable/features/reminders/presentation/reminder_details_form.dart';
import 'package:control_contable/features/reminders/presentation/upcoming_due_dates_section.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The loaded reminders screen, one column: availability or permission
/// notice, the opt-in switch, the details while on, the next due dates.
class RemindersView extends ConsumerStatefulWidget {
  /// The view of [settings] as the server holds them.
  const new({required this.settings, super.key});

  /// The saved settings.
  final ReminderSettings settings;

  @override
  ConsumerState<RemindersView> createState() => _RemindersViewState();
}

class _RemindersViewState extends ConsumerState<RemindersView> {
  late final AppLifecycleListener _lifecycle;
  bool _switching = false;
  bool _deniedNow = false;
  String? _failure;

  @override
  void initState() {
    super.initState();
    // Back from the system settings: the permission may have changed.
    _lifecycle = AppLifecycleListener(
      onResume: () => ref.invalidate(pushPermissionProvider),
    );
  }

  @override
  void dispose() {
    _lifecycle.dispose();
    super.dispose();
  }

  Future<void> _onSwitch({required bool on}) async {
    if (_switching) return;
    final l10n = AppLocalizations.of(context);
    final notifier = ref.read(reminderSettingsProvider.notifier);
    setState(() {
      _switching = true;
      _failure = null;
    });
    try {
      if (on) {
        final outcome = await notifier.turnOn();
        _deniedNow = outcome == TurnOnOutcome.permissionDenied;
        if (outcome == TurnOnOutcome.enabled) {
          ref.invalidate(pushPermissionProvider);
        }
      } else {
        await notifier.turnOff();
      }
    } on Exception {
      _failure = on ? l10n.remindersTurnOnFailed : l10n.remindersTurnOffFailed;
    } finally {
      if (mounted) setState(() => _switching = false);
    }
  }

  bool get _blockedInOs {
    final permission = ref.watch(pushPermissionProvider).value;
    return _deniedNow ||
        (widget.settings.enabled && permission == PushPermission.denied);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final available = ref.watch(pushMessagingProvider).isAvailable;
    final enabled = widget.settings.enabled;
    return ListView(
      padding: const EdgeInsets.symmetric(vertical: AppSpacing.md),
      children: [
        if (!available)
          _Padded(FormNotice(message: l10n.remindersUnavailable))
        else if (_blockedInOs)
          const _Padded(PushDeniedNotice()),
        if (_failure case final failure?) _Padded(FormNotice(message: failure)),
        SwitchListTile(
          title: Text(l10n.remindersEnabledLabel),
          subtitle: Text(l10n.remindersEnabledHelp),
          value: enabled,
          onChanged: (on) => _onSwitch(on: on),
        ),
        if (enabled) ReminderDetailsForm(settings: widget.settings),
        const SizedBox(height: AppSpacing.xl),
        const _Padded(UpcomingDueDatesSection()),
      ],
    );
  }
}

class _Padded extends StatelessWidget {
  const new(this.child);

  final Widget child;

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.fromLTRB(
      AppSpacing.gutter,
      0,
      AppSpacing.gutter,
      AppSpacing.md,
    ),
    child: child,
  );
}
