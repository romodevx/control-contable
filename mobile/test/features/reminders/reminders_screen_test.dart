import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/reminders/domain/push_token.dart';
import 'package:control_contable/features/reminders/presentation/reminders_screen.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'pump_reminders.dart';
import 'reminders_fixtures.dart';

const String _master = 'Recibir recordatorios de pago';
const String _earlyLabel = 'Recordatorio anticipado';
const String _dueLabel = 'Avisarme el día del vencimiento';
const String _timeLabel = 'Hora del aviso';
const String _amounts =
    'Mostrar montos en las notificaciones (se verán en la pantalla de '
    'bloqueo)';
const String _saved = 'Guardado.';
const String _saveFailed =
    'No se pudo guardar el cambio; se restauró el valor anterior. Inténtelo '
    'de nuevo.';
const String _noneSelected =
    'Elija al menos un recordatorio: el anticipado o el del día del '
    'vencimiento.';
const String _openSettings = 'Abrir ajustes del sistema';

Finder _switchOf(String label) => find.widgetWithText(SwitchListTile, label);

bool _isOn(WidgetTester tester, String label) =>
    tester.widget<SwitchListTile>(_switchOf(label)).value;

List<String> _announced(WidgetTester tester) => [
  for (final announcement in tester.takeAnnouncements()) announcement.message,
];

Future<void> _chooseEarlyDay(WidgetTester tester, String option) async {
  await tester.tap(find.byType(DropdownButton<int?>));
  await tester.pumpAndSettle();
  await tester.tap(find.text(option).last);
  await tester.pumpAndSettle();
}

void _tall(WidgetTester tester) {
  tester.view
    ..physicalSize = const Size(400, 2400)
    ..devicePixelRatio = 1;
  addTearDown(tester.view.reset);
}

void main() {
  late RemindersHarness harness;

  setUp(() => harness = RemindersHarness());

  group('module_reminders', () {
    testWidgets('OFF: no entry on Home and the location is not found', (
      tester,
    ) async {
      await harness.pump(
        tester,
        modules: remindersModules.difference({AppModule.reminders}),
      );

      expect(find.byType(RemindersScreen), findsNothing);
      expect(harness.repository.puts, isEmpty);
      expect(harness.messaging.prompts, 0);
    });

    testWidgets('OFF: Home has no "Recordatorios" entry', (tester) async {
      await harness.pump(
        tester,
        modules: remindersModules.difference({AppModule.reminders}),
        location: null,
      );

      expect(find.byTooltip('Recordatorios'), findsNothing);
    });

    testWidgets('ON: Home offers "Recordatorios", which opens the screen', (
      tester,
    ) async {
      await harness.pump(tester, location: null);

      await tester.tap(find.byTooltip('Recordatorios'));
      await tester.pumpAndSettle();

      expect(harness.location, '/recordatorios');
      expect(find.byType(RemindersScreen), findsOneWidget);
    });
  });

  group('opt in (US-R1)', () {
    testWidgets('never asked: off, details hidden, no OS prompt on visit', (
      tester,
    ) async {
      await harness.pump(tester);

      expect(_isOn(tester, _master), isFalse);
      expect(find.text(_earlyLabel), findsNothing);
      expect(harness.messaging.prompts, 0, reason: 'never at launch/visit');
    });

    testWidgets('turning it on prompts, registers and saves', (tester) async {
      _tall(tester);
      await harness.pump(tester);

      await tester.tap(_switchOf(_master));
      await tester.pumpAndSettle();

      expect(harness.messaging.prompts, 1);
      expect(
        harness.repository.registered.single.installationId,
        testInstallationId,
      );
      expect(harness.repository.puts.single.enabled, isTrue);
      expect(_isOn(tester, _master), isTrue);
      expect(find.text(_earlyLabel), findsOneWidget);
      expect(find.text('El día 14 del mes de pago'), findsOneWidget);
      expect(_isOn(tester, _dueLabel), isTrue);
      expect(find.textContaining('10:00'), findsOneWidget);
      expect(_isOn(tester, _amounts), isFalse, reason: 'default OFF');
    });

    testWidgets('denied: stays off and explains how to allow them', (
      tester,
    ) async {
      _tall(tester);
      harness.messaging.answer = PushPermission.denied;
      await harness.pump(tester);

      await tester.tap(_switchOf(_master));
      await tester.pumpAndSettle();

      expect(_isOn(tester, _master), isFalse);
      expect(harness.repository.puts, isEmpty);
      expect(find.textContaining('bloqueadas'), findsOneWidget);

      await tester.tap(find.widgetWithText(OutlinedButton, _openSettings));
      await tester.pumpAndSettle();
      expect(harness.settingsOpened, 1);
    });

    testWidgets('enabled but blocked in the OS: explained, never re-prompted', (
      tester,
    ) async {
      _tall(tester);
      harness.repository.settings = defaultSettings.copyWith(enabled: true);
      harness.messaging.status = PushPermission.denied;
      await harness.pump(tester);

      expect(find.textContaining('bloqueadas'), findsOneWidget);
      expect(find.text(_openSettings), findsOneWidget);
      expect(harness.messaging.prompts, 0);
    });

    testWidgets('push not available in this build: said, nothing asked', (
      tester,
    ) async {
      harness.messaging.isAvailable = false;
      await harness.pump(tester);

      expect(find.textContaining('no disponibles'), findsOneWidget);

      await tester.tap(_switchOf(_master));
      await tester.pumpAndSettle();

      expect(_isOn(tester, _master), isFalse);
      expect(harness.messaging.prompts, 0);
      expect(harness.repository.registered, isEmpty);
    });

    testWidgets('a failed registration keeps it off and says so', (
      tester,
    ) async {
      harness.repository.registerError = const ApiUnreachable();
      await harness.pump(tester);

      await tester.tap(_switchOf(_master));
      await tester.pumpAndSettle();

      expect(_isOn(tester, _master), isFalse);
      expect(harness.repository.puts, isEmpty);
      expect(
        find.text(
          'No se pudieron activar los recordatorios. Inténtelo de '
          'nuevo.',
        ),
        findsOneWidget,
      );
    });

    testWidgets('turning it off saves and unregisters (US-R8)', (tester) async {
      _tall(tester);
      harness.repository.settings = defaultSettings.copyWith(enabled: true);
      harness.messaging.status = PushPermission.granted;
      harness.store.registered = true;
      await harness.pump(tester);

      await tester.tap(_switchOf(_master));
      await tester.pumpAndSettle();

      expect(harness.repository.puts.single.enabled, isFalse);
      expect(harness.repository.unregistered, [testInstallationId]);
      expect(find.text(_earlyLabel), findsNothing);
    });
  });

  group('the details (enabled)', () {
    setUp(() {
      harness.repository.settings = defaultSettings.copyWith(enabled: true);
      harness.messaging.status = PushPermission.granted;
    });

    testWidgets('amounts toggle hidden while module_tax_engine is OFF', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(
        tester,
        modules: remindersModules.difference({AppModule.taxEngine}),
      );

      expect(find.text(_earlyLabel), findsOneWidget);
      expect(_switchOf(_amounts), findsNothing);
    });

    testWidgets('lock-screen privacy is explained next to the toggle', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(tester);

      expect(find.textContaining('pantalla de bloqueo'), findsWidgets);
      expect(find.textContaining('nunca su RFC'), findsOneWidget);
    });

    testWidgets('autosave: no save button, no primary needed (§7)', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(tester);

      expect(find.byType(FilledButton), findsNothing);
      expect(find.text('Guardar cambios'), findsNothing);
    });

    testWidgets('each change saves the whole object at once and says so', (
      tester,
    ) async {
      _tall(tester);
      final handle = tester.ensureSemantics();
      await harness.pump(tester);
      tester.takeAnnouncements();

      await tester.tap(_switchOf(_amounts));
      await tester.pumpAndSettle();

      expect(
        harness.repository.puts.single,
        defaultSettings.copyWith(enabled: true, showAmounts: true),
      );
      expect(_isOn(tester, _amounts), isTrue);
      expect(find.text(_saved), findsOneWidget);
      expect(_announced(tester), [_saved]);

      await _chooseEarlyDay(tester, 'Sin recordatorio anticipado');

      expect(harness.repository.puts.last.earlyDay, isNull);
      expect(_announced(tester), [_saved], reason: 'every save, not once');
      handle.dispose();
    });

    testWidgets('a failed change reverts the control and says so', (
      tester,
    ) async {
      _tall(tester);
      final handle = tester.ensureSemantics();
      await harness.pump(tester);
      tester.takeAnnouncements();
      harness.repository.putReply = const ApiUnreachable();

      await tester.tap(_switchOf(_amounts));
      await tester.pumpAndSettle();

      expect(_isOn(tester, _amounts), isFalse, reason: 'reverted');
      expect(find.text(_saveFailed), findsOneWidget);
      expect(find.text(_saved), findsNothing);
      expect(_announced(tester), [_saveFailed]);
      handle.dispose();
    });

    testWidgets('a failed early-day change shows the saved day again', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(tester);
      harness.repository.putReply = const ApiUnreachable();

      await _chooseEarlyDay(tester, 'El día 10 del mes de pago');

      expect(find.text('El día 14 del mes de pago'), findsOneWidget);
      expect(find.text('El día 10 del mes de pago'), findsNothing);
      expect(find.text(_saveFailed), findsOneWidget);
    });

    testWidgets('leaving nothing to send is refused: no PUT, switch stays', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(tester);

      await _chooseEarlyDay(tester, 'Sin recordatorio anticipado');
      await tester.tap(_switchOf(_dueLabel));
      await tester.pumpAndSettle();

      expect(harness.repository.puts, hasLength(1), reason: 'only the day');
      expect(_isOn(tester, _dueLabel), isTrue);
      expect(find.text(_noneSelected), findsOneWidget);
    });

    testWidgets('a later valid change clears the error', (tester) async {
      _tall(tester);
      harness.repository.putReply = const ApiUnreachable();
      await harness.pump(tester);

      await tester.tap(_switchOf(_amounts));
      await tester.pumpAndSettle();
      expect(find.text(_saveFailed), findsOneWidget);

      await tester.tap(_switchOf(_amounts));
      await tester.pumpAndSettle();

      expect(find.text(_saveFailed), findsNothing);
      expect(find.text(_saved), findsOneWidget);
    });

    testWidgets('the time uses the platform picker', (tester) async {
      _tall(tester);
      await harness.pump(tester);

      await tester.tap(find.bySemanticsLabel(RegExp(_timeLabel)));
      await tester.pumpAndSettle();

      expect(find.byType(TimePickerDialog), findsOneWidget);
    });
  });

  group('upcoming due dates', () {
    testWidgets('the next three, moved ones marked, subject to confirmation', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(tester);

      expect(find.text('Próximos vencimientos'), findsOneWidget);
      expect(find.text('lunes, 18 de mayo de 2026'), findsOneWidget);
      expect(find.text('Se recorre por fin de semana'), findsOneWidget);
      expect(find.text('miércoles, 17 de junio de 2026'), findsOneWidget);
      expect(find.text('viernes, 17 de julio de 2026'), findsOneWidget);
      expect(find.text('Pago mensual de abril de 2026'), findsOneWidget);
      expect(find.text('Fechas sujetas a confirmación.'), findsOneWidget);
      expect(harness.repository.calendarYears, [2026]);
    });

    testWidgets('late in the year it reads the next year too', (tester) async {
      _tall(tester);
      harness.now = DateTime.utc(2026, 12, 20, 18);
      harness.repository.calendars[2027] = taxCalendar(2027);
      await harness.pump(tester);

      expect(harness.repository.calendarYears, [2026, 2027]);
      expect(find.text('miércoles, 17 de febrero de 2027'), findsOneWidget);
      expect(find.text('domingo, 17 de enero de 2027'), findsOneWidget);
    });

    testWidgets("in January, last year's December period is still due", (
      tester,
    ) async {
      _tall(tester);
      harness.now = DateTime.utc(2027, 1, 5, 18);
      harness.repository.calendars[2027] = taxCalendar(2027);
      await harness.pump(tester);

      expect(harness.repository.calendarYears, [2026, 2027]);
      expect(find.text('domingo, 17 de enero de 2027'), findsOneWidget);
      expect(find.text('Pago mensual de diciembre de 2026'), findsOneWidget);
    });

    testWidgets('an unseeded next year still shows this year', (tester) async {
      _tall(tester);
      harness.now = DateTime.utc(2026, 12, 20, 18);
      harness.repository.calendarError = const ApiRejected(code: null);
      await harness.pump(tester);

      expect(find.text('domingo, 17 de enero de 2027'), findsOneWidget);
    });
  });

  group('states and accessibility', () {
    testWidgets('a failed read offers a retry', (tester) async {
      harness.repository.getError = const ApiUnreachable();
      await harness.pump(tester);

      expect(
        find.text('No se pudieron cargar sus recordatorios.'),
        findsOneWidget,
      );
      await tester.tap(find.text('Reintentar').first);
      await tester.pumpAndSettle();

      expect(_switchOf(_master), findsOneWidget);
    });

    testWidgets('targets and labels meet the guidelines', (tester) async {
      _tall(tester);
      harness.repository.settings = defaultSettings.copyWith(enabled: true);
      harness.messaging.status = PushPermission.granted;
      final handle = tester.ensureSemantics();
      await harness.pump(tester);

      await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
      await expectLater(tester, meetsGuideline(iOSTapTargetGuideline));
      await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
      await expectLater(tester, meetsGuideline(textContrastGuideline));
      handle.dispose();
    });

    testWidgets('200 % text lays out without overflow', (tester) async {
      harness.repository.settings = defaultSettings.copyWith(enabled: true);
      harness.messaging.status = PushPermission.granted;
      tester.platformDispatcher.textScaleFactorTestValue = 2;
      addTearDown(tester.platformDispatcher.clearTextScaleFactorTestValue);
      await harness.pump(tester);

      expect(tester.takeException(), isNull);
      await tester.scrollUntilVisible(_switchOf(_amounts), 200);
      expect(_switchOf(_amounts), findsOneWidget);
    });
  });
}
