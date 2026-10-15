import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/features/auth/application/sign_out.dart';
import 'package:control_contable/features/reminders/domain/push_token.dart';
import 'package:flutter_test/flutter_test.dart';

import 'pump_reminders.dart';
import 'reminders_fixtures.dart';

const String _link = 'controlcontable://estimate?period=2026-04';

void main() {
  late RemindersHarness harness;

  setUp(() => harness = RemindersHarness());

  group('notification tap (deep link, §2.5)', () {
    testWidgets('opens the estimate of the period', (tester) async {
      await harness.pump(tester, location: null);

      harness.messaging.tap(_link);
      await tester.pumpAndSettle();

      expect(harness.location, '/estimado?periodo=2026-04');
    });

    testWidgets('with the engine OFF it opens Home', (tester) async {
      await harness.pump(
        tester,
        modules: remindersModules.difference({AppModule.taxEngine}),
      );

      harness.messaging.tap(_link);
      await tester.pumpAndSettle();

      expect(harness.location, '/');
    });

    testWidgets('a malformed link opens Home', (tester) async {
      await harness.pump(tester);

      harness.messaging.tap('controlcontable://estimate?period=abril');
      await tester.pumpAndSettle();

      expect(harness.location, '/');
    });

    testWidgets('the notification that launched the app is followed', (
      tester,
    ) async {
      harness.messaging.launchLink = _link;
      await harness.pump(tester, location: null);

      expect(harness.location, '/estimado?periodo=2026-04');
    });
  });

  group('device registration after sign-in (§2.3)', () {
    const signedInModules = {...remindersModules, AppModule.auth};

    setUp(() => harness.messaging.status = PushPermission.granted);

    testWidgets('enabled + permission: registers once on sign-in', (
      tester,
    ) async {
      harness.repository.settings = defaultSettings.copyWith(enabled: true);
      await harness.pump(
        tester,
        modules: signedInModules,
        signedIn: true,
        location: null,
      );

      expect(harness.repository.registered, hasLength(1));
      expect(harness.messaging.prompts, 0, reason: 'never prompts here');
    });

    testWidgets('reminders off: nothing registered', (tester) async {
      await harness.pump(
        tester,
        modules: signedInModules,
        signedIn: true,
        location: null,
      );

      expect(harness.repository.registered, isEmpty);
    });

    testWidgets('no OS permission: settings not even read', (tester) async {
      harness.messaging.status = PushPermission.notDetermined;
      harness.repository.settings = defaultSettings.copyWith(enabled: true);
      await harness.pump(
        tester,
        modules: signedInModules,
        signedIn: true,
        location: null,
      );

      expect(harness.repository.registered, isEmpty);
      expect(harness.messaging.prompts, 0);
    });

    testWidgets('module_reminders OFF: nothing happens', (tester) async {
      harness.repository.settings = defaultSettings.copyWith(enabled: true);
      await harness.pump(
        tester,
        modules: signedInModules.difference({AppModule.reminders}),
        signedIn: true,
        location: null,
      );

      expect(harness.repository.registered, isEmpty);
    });

    testWidgets('signed out: nothing happens', (tester) async {
      harness.repository.settings = defaultSettings.copyWith(enabled: true);
      await harness.pump(tester, modules: signedInModules, location: null);

      expect(harness.repository.registered, isEmpty);
    });

    testWidgets('a rotated token re-registers the device', (tester) async {
      harness.repository.settings = defaultSettings.copyWith(enabled: true);
      await harness.pump(
        tester,
        modules: signedInModules,
        signedIn: true,
        location: null,
      );

      harness.messaging.rotate(const PushToken('fcm-token-2'));
      await tester.pumpAndSettle();

      expect(harness.repository.registered.map((call) => call.token), [
        const PushToken('fcm-token-1'),
        const PushToken('fcm-token-2'),
      ]);
    });

    testWidgets('sign-out unregisters before revoking (US-R8)', (tester) async {
      harness.repository.settings = defaultSettings.copyWith(enabled: true);
      final order = <String>[];
      harness.auth.onLogout = () => order.add(
        'revoke after ${harness.repository.unregistered.length} delete',
      );
      await harness.pump(
        tester,
        modules: signedInModules,
        signedIn: true,
        location: null,
      );

      await harness.container.read(signOutProvider).call();
      await tester.pumpAndSettle();

      expect(harness.repository.unregistered, [testInstallationId]);
      expect(order, ['revoke after 1 delete']);
    });
  });
}
