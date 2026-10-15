import 'package:control_contable/api/generated/models/reminder_settings.dart';
import 'package:control_contable/features/reminders/application/push_device.dart';
import 'package:control_contable/features/reminders/application/push_messaging.dart';
import 'package:control_contable/features/reminders/data/reminders_repository.dart';
import 'package:control_contable/features/reminders/domain/push_token.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// How turning reminders on ended.
enum TurnOnOutcome {
  /// Permission granted, device registered, settings saved enabled.
  enabled,

  /// The OS permission was refused: nothing registered or saved.
  permissionDenied,

  /// Push does not work in this build: nothing asked, registered or saved.
  unavailable,
}

/// The user's reminder settings as the server holds them, and the three
/// ways to change them (MODULE-reminders §2.2, US-R1, US-R8). Read on each
/// visit; a failed read shows a retry, never an automatic one.
class ReminderSettingsNotifier extends AsyncNotifier<ReminderSettings> {
  @override
  Future<ReminderSettings> build() =>
      ref.read(remindersRepositoryProvider).getSettings();

  /// Opt-in: the OS prompt first (only here, never at launch), then the
  /// device registration, then the save. Throws an `ApiFailure` or
  /// [PushUnavailable] when a step fails after the permission.
  Future<TurnOnOutcome> turnOn() async {
    final messaging = ref.read(pushMessagingProvider);
    if (!messaging.isAvailable) return TurnOnOutcome.unavailable;
    final permission = await messaging.requestPermission();
    if (permission != PushPermission.granted) {
      return TurnOnOutcome.permissionDenied;
    }
    await ref.read(pushDeviceProvider).register();
    await save(_withAReminder(state.requireValue).copyWith(enabled: true));
    return TurnOnOutcome.enabled;
  }

  /// Opt-out: saves disabled, then removes the device (best effort; a
  /// failure is thrown after the save went through).
  Future<void> turnOff() async {
    await save(state.requireValue.copyWith(enabled: false));
    await ref.read(pushDeviceProvider).unregister();
  }

  /// Replaces the whole object with [settings].
  Future<void> save(ReminderSettings settings) async {
    final saved = await ref
        .read(remindersRepositoryProvider)
        .putSettings(settings);
    if (ref.mounted) state = AsyncData(saved);
  }

  /// [settings], with the due-date reminder on if none is selected: "on"
  /// with nothing to send is a `422 no_reminder_selected`.
  static ReminderSettings _withAReminder(ReminderSettings settings) =>
      settings.earlyDay == null && !settings.onDueDate
      ? settings.copyWith(onDueDate: true)
      : settings;
}

/// See [ReminderSettingsNotifier].
final AsyncNotifierProvider<ReminderSettingsNotifier, ReminderSettings>
reminderSettingsProvider =
    AsyncNotifierProvider.autoDispose<
      ReminderSettingsNotifier,
      ReminderSettings
    >(
      ReminderSettingsNotifier.new,
      name: 'reminderSettingsProvider',
      retry: _noAutomaticRetry,
    );

Duration? _noAutomaticRetry(int retryCount, Object error) => null;
