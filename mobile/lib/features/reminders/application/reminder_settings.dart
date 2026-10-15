import 'package:control_contable/api/generated/models/reminder_settings.dart';
import 'package:control_contable/core/diagnostics/error_reporter.dart';
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

/// How an autosaved [ReminderSettingsNotifier.change] ended.
enum ChangeOutcome {
  /// The server holds it.
  saved,

  /// A later change replaced it before it was saved, or a failure dropped
  /// it; the later change (or the failure) reports for both.
  superseded,

  /// It would leave reminders on with nothing to send (`422
  /// no_reminder_selected`): refused locally, nothing sent.
  noReminderSelected,
}

/// The user's reminder settings as the server holds them, and the ways to
/// change them (MODULE-reminders §2.2, §7, US-R1, US-R8). Read on each
/// visit; a failed read shows a retry, never an automatic one.
class ReminderSettingsNotifier extends AsyncNotifier<ReminderSettings> {
  /// What the server last confirmed: a failed change reverts to it.
  late ReminderSettings _confirmed;

  /// Autosaves run one at a time, in the order the user made them.
  Future<void> _queue = Future.value();

  /// Bumped by each [build]: a late reply of an older run is ignored.
  int _generation = 0;

  @override
  Future<ReminderSettings> build() async {
    _generation++;
    _queue = Future.value();
    return _confirmed = await ref
        .read(remindersRepositoryProvider)
        .getSettings();
  }

  /// Autosave of one detail change (§7: no save button). Shows [next] at
  /// once, then replaces the whole object on the server, after any change
  /// still saving. A failure reverts to what the server holds, drops the
  /// changes queued behind it, and is rethrown (`ApiFailure`).
  Future<ChangeOutcome> change(ReminderSettings next) {
    if (next.enabled && !_hasAReminder(next)) {
      return Future.value(ChangeOutcome.noReminderSelected);
    }
    state = AsyncData(next);
    final generation = _generation;
    final run = _queue.then((_) => _saveIfLatest(next, generation));
    _queue = run.then<void>((_) {}, onError: (Object _) {});
    return run;
  }

  Future<ChangeOutcome> _saveIfLatest(
    ReminderSettings next,
    int generation,
  ) async {
    bool current() =>
        ref.mounted && generation == _generation && state.value == next;
    if (!current()) return ChangeOutcome.superseded;
    try {
      final saved = await ref
          .read(remindersRepositoryProvider)
          .putSettings(next);
      if (!ref.mounted || generation != _generation) {
        return ChangeOutcome.superseded;
      }
      _confirmed = saved;
      if (!current()) return ChangeOutcome.superseded;
      state = AsyncData(saved);
      return ChangeOutcome.saved;
    } on Exception {
      if (ref.mounted && generation == _generation) {
        state = AsyncData(_confirmed);
      }
      rethrow;
    }
  }

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

  /// Opt-out: saves disabled (throws when that fails), then removes the
  /// device, best effort: with reminders off nothing is sent to it anyway.
  Future<void> turnOff() async {
    await save(state.requireValue.copyWith(enabled: false));
    try {
      await ref.read(pushDeviceProvider).unregister();
    } on Exception catch (error) {
      ref.read(errorReporterProvider).report('reminders.unregister', error);
    }
  }

  /// Replaces the whole object with [settings], showing it once saved (the
  /// opt-in switch: it moves only when the server holds it).
  Future<void> save(ReminderSettings settings) async {
    final generation = _generation;
    final saved = await ref
        .read(remindersRepositoryProvider)
        .putSettings(settings);
    if (!ref.mounted || generation != _generation) return;
    state = AsyncData(_confirmed = saved);
  }

  static bool _hasAReminder(ReminderSettings settings) =>
      settings.earlyDay != null || settings.onDueDate;

  /// [settings], with the due-date reminder on if none is selected: "on"
  /// with nothing to send is a `422 no_reminder_selected`.
  static ReminderSettings _withAReminder(ReminderSettings settings) =>
      _hasAReminder(settings) ? settings : settings.copyWith(onDueDate: true);
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

/// The OS notification permission, read without prompting; re-read when
/// the app resumes (the user may come back from the system settings).
final FutureProvider<PushPermission> pushPermissionProvider =
    FutureProvider.autoDispose<PushPermission>(
      (ref) => ref.read(pushMessagingProvider).permission(),
      name: 'pushPermissionProvider',
      retry: _noAutomaticRetry,
    );
