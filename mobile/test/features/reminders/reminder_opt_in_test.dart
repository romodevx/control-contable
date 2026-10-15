import 'dart:async';

import 'package:control_contable/api/generated/models/device_platform.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/reminders/application/push_device.dart';
import 'package:control_contable/features/reminders/application/push_messaging.dart';
import 'package:control_contable/features/reminders/application/reminder_settings.dart';
import 'package:control_contable/features/reminders/data/installation_store.dart';
import 'package:control_contable/features/reminders/data/reminders_repository.dart';
import 'package:control_contable/features/reminders/domain/push_token.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/recording_adapter.dart';
import 'reminders_fakes.dart';
import 'reminders_fixtures.dart';

const String _installationId = '01920000-0000-7000-8000-00000000000a';

void main() {
  late FakePushMessaging messaging;
  late FakeRemindersRepository repository;
  late InMemoryInstallationStore store;
  late PushDevice device;

  setUp(() {
    messaging = FakePushMessaging();
    repository = FakeRemindersRepository();
    store = InMemoryInstallationStore(_installationId);
    device = PushDevice(
      messaging: messaging,
      repository: repository,
      store: store,
      platform: DevicePlatform.android,
    );
  });

  group('PushDevice', () {
    test('register sends the installation id and the current token', () async {
      await device.register();

      expect(repository.registered.single, (
        installationId: _installationId,
        platform: DevicePlatform.android,
        token: const PushToken('fcm-token-1'),
      ));
      expect(await store.isRegistered(), isTrue);
    });

    test(
      'no token from the provider is PushUnavailable, nothing sent',
      () async {
        messaging.currentToken = null;

        expect(await failureOf(device.register), isA<PushUnavailable>());
        expect(repository.registered, isEmpty);
      },
    );

    test('a rotated token re-registers only a registered device', () async {
      await device.onTokenRefreshed(const PushToken('fcm-token-2'));
      expect(repository.registered, isEmpty);

      await device.register();
      await device.onTokenRefreshed(const PushToken('fcm-token-2'));

      expect(repository.registered.last.token, const PushToken('fcm-token-2'));
      expect(repository.registered.last.installationId, _installationId);
    });

    test('unregister deletes a registered device once', () async {
      await device.unregister();
      expect(repository.unregistered, isEmpty, reason: 'never registered');

      await device.register();
      await device.unregister();
      await device.unregister();

      expect(repository.unregistered, [_installationId]);
      expect(await store.isRegistered(), isFalse);
    });

    test('a 404 on unregister counts as done', () async {
      await device.register();
      repository.unregisterError = const ApiNotFound();

      await device.unregister();

      expect(await store.isRegistered(), isFalse);
    });

    test('another failure is thrown, and the marker still clears', () async {
      await device.register();
      repository.unregisterError = const ApiUnreachable();

      expect(await failureOf(device.unregister), isA<ApiUnreachable>());
      expect(await store.isRegistered(), isFalse);
    });
  });

  group('turning reminders on and off (US-R1, US-R8)', () {
    late ProviderContainer container;

    setUp(() {
      container = ProviderContainer(
        overrides: [
          pushMessagingProvider.overrideWithValue(messaging),
          remindersRepositoryProvider.overrideWithValue(repository),
          installationStoreProvider.overrideWithValue(store),
          devicePlatformProvider.overrideWithValue(DevicePlatform.android),
        ],
      );
      addTearDown(container.dispose);
      container.listen(reminderSettingsProvider, (_, _) {});
    });

    ReminderSettingsNotifier notifier() =>
        container.read(reminderSettingsProvider.notifier);

    test('reads the defaults', () async {
      expect(
        await container.read(reminderSettingsProvider.future),
        defaultSettings,
      );
    });

    test('on: prompts, registers, then saves enabled', () async {
      await container.read(reminderSettingsProvider.future);

      final outcome = await notifier().turnOn();

      expect(outcome, TurnOnOutcome.enabled);
      expect(messaging.prompts, 1);
      expect(repository.registered, hasLength(1));
      expect(repository.puts.single, defaultSettings.copyWith(enabled: true));
      expect(container.read(reminderSettingsProvider).value?.enabled, isTrue);
    });

    test('on, permission denied: nothing registered, nothing saved', () async {
      await container.read(reminderSettingsProvider.future);
      messaging.answer = PushPermission.denied;

      expect(await notifier().turnOn(), TurnOnOutcome.permissionDenied);
      expect(repository.registered, isEmpty);
      expect(repository.puts, isEmpty);
    });

    test('on, push not available in this build: skipped', () async {
      await container.read(reminderSettingsProvider.future);
      messaging.isAvailable = false;

      expect(await notifier().turnOn(), TurnOnOutcome.unavailable);
      expect(messaging.prompts, 0);
      expect(repository.registered, isEmpty);
      expect(repository.puts, isEmpty);
    });

    test('on with no reminder selected selects the due-date one', () async {
      repository.settings = defaultSettings.copyWith(
        earlyDay: null,
        onDueDate: false,
      );
      await container.read(reminderSettingsProvider.future);

      await notifier().turnOn();

      expect(repository.puts.single.onDueDate, isTrue);
      expect(repository.puts.single.enabled, isTrue);
    });

    test('off: saves disabled, then unregisters the device', () async {
      await container.read(reminderSettingsProvider.future);
      await notifier().turnOn();

      await notifier().turnOff();

      expect(repository.puts.last.enabled, isFalse);
      expect(repository.unregistered, [_installationId]);
    });

    test('change shows the new value at once and saves the whole object', () async {
      repository.settings = defaultSettings.copyWith(enabled: true);
      await container.read(reminderSettingsProvider.future);
      final pending = Completer<void>();
      repository.putReply = pending;
      final next = repository.settings.copyWith(showAmounts: true);

      final saving = notifier().change(next);

      expect(container.read(reminderSettingsProvider).value, next);
      pending.complete();
      expect(await saving, ChangeOutcome.saved);
      expect(repository.puts.single, next);
      expect(container.read(reminderSettingsProvider).value, next);
    });

    test('a failed change reverts to what the server holds and throws', () async {
      final saved = defaultSettings.copyWith(enabled: true);
      repository.settings = saved;
      await container.read(reminderSettingsProvider.future);
      repository.putReply = const ApiUnreachable();

      await expectLater(
        notifier().change(saved.copyWith(sendTime: '08:00')),
        throwsA(isA<ApiUnreachable>()),
      );

      expect(container.read(reminderSettingsProvider).value, saved);
    });

    test('a change leaving nothing to send is refused locally, no PUT', () async {
      final saved = defaultSettings.copyWith(enabled: true, earlyDay: null);
      repository.settings = saved;
      await container.read(reminderSettingsProvider.future);

      final outcome = await notifier().change(saved.copyWith(onDueDate: false));

      expect(outcome, ChangeOutcome.noReminderSelected);
      expect(repository.puts, isEmpty);
      expect(container.read(reminderSettingsProvider).value, saved);
    });

    test('rapid changes save in order and the last one wins', () async {
      final saved = defaultSettings.copyWith(enabled: true);
      repository.settings = saved;
      await container.read(reminderSettingsProvider.future);
      final pending = Completer<void>();
      repository.putReply = pending;
      final first = saved.copyWith(sendTime: '08:00');
      final second = first.copyWith(showAmounts: true);

      final firstSave = notifier().change(first);
      final secondSave = notifier().change(second);
      pending.complete();

      expect(await firstSave, ChangeOutcome.superseded);
      expect(await secondSave, ChangeOutcome.saved);
      expect(repository.puts, [first, second]);
      expect(container.read(reminderSettingsProvider).value, second);
    });

    test('a failure drops the queued change too: nothing half-applied', () async {
      final saved = defaultSettings.copyWith(enabled: true);
      repository.settings = saved;
      await container.read(reminderSettingsProvider.future);
      final pending = Completer<void>();
      repository.putReply = pending;
      final first = saved.copyWith(sendTime: '08:00');

      final firstSave = notifier().change(first);
      final secondSave = notifier().change(first.copyWith(showAmounts: true));
      pending.completeError(const ApiUnreachable());

      await expectLater(firstSave, throwsA(isA<ApiUnreachable>()));
      expect(await secondSave, ChangeOutcome.superseded);
      expect(repository.puts, [first]);
      expect(container.read(reminderSettingsProvider).value, saved);
    });

    test('save replaces the whole object', () async {
      await container.read(reminderSettingsProvider.future);
      final draft = defaultSettings.copyWith(enabled: true, sendTime: '08:00');

      await notifier().save(draft);

      expect(repository.puts.single, draft);
      expect(container.read(reminderSettingsProvider).value, draft);
    });
  });
}
