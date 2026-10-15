import 'package:control_contable/core/diagnostics/error_reporter.dart';
import 'package:control_contable/features/reminders/application/push_device.dart';
import 'package:control_contable/features/reminders/application/push_messaging.dart';
import 'package:control_contable/features/reminders/data/reminders_repository.dart';
import 'package:control_contable/features/reminders/domain/push_token.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Keeps this installation registered while the user is signed in with
/// reminders on (MODULE-reminders §2.3): after every sign-in (a cold start
/// with a kept session included) and whenever FCM rotates the token. It
/// never prompts for the permission: that happens only when the user turns
/// reminders on. Every failure is reported and absorbed (best effort: the
/// next sign-in, or the claim, catches up).
class ReminderPushSync {
  /// The sync over its collaborators.
  new({
    required this._messaging,
    required this._repository,
    required this._device,
    required this._reporter,
  });

  final PushMessaging _messaging;
  final RemindersRepository _repository;
  final PushDevice _device;
  final ErrorReporter _reporter;

  /// The user just signed in with `module_reminders` on: registers when
  /// push works, the OS allows it and the user's reminders are on. A
  /// registration still marked from before (a sign-out whose delete
  /// failed, M9 security gate F1) is claimed for this user either way, and
  /// removed when this user's reminders are off, so the previous user's
  /// reminders never reach this phone.
  Future<void> afterSignIn() => _absorbed('reminders.sync', () async {
    if (!_messaging.isAvailable) return;
    final granted = await _messaging.permission() == PushPermission.granted;
    final leftover = await _device.registeredId() != null;
    if (!granted && !leftover) return;
    final settings = await _repository.getSettings();
    if (settings.enabled && granted) {
      await _device.register();
    } else if (!settings.enabled && leftover) {
      await _device.evict();
    }
  });

  /// FCM rotated the token: re-registers a registered installation.
  Future<void> onTokenRefreshed(PushToken token) => _absorbed(
    'reminders.tokenRefresh',
    () => _device.onTokenRefreshed(token),
  );

  Future<void> _absorbed(String operation, Future<void> Function() run) async {
    try {
      await run();
    } on Exception catch (error) {
      _reporter.report(operation, error);
    }
  }
}

/// See [ReminderPushSync].
final reminderPushSyncProvider = Provider<ReminderPushSync>(
  (ref) => ReminderPushSync(
    messaging: ref.watch(pushMessagingProvider),
    repository: ref.watch(remindersRepositoryProvider),
    device: ref.watch(pushDeviceProvider),
    reporter: ref.watch(errorReporterProvider),
  ),
  name: 'reminderPushSyncProvider',
);
