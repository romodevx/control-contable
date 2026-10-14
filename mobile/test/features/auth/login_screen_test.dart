import 'dart:async';

import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/auth/auth_routes.dart';
import 'package:control_contable/features/auth/presentation/login_screen.dart';
import 'package:control_contable/features/home/presentation/home_screen.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/synthetic_password.dart';
import 'pump_auth.dart';

const _title = 'Iniciar sesión';
const _email = 'Correo electrónico';
const _password = 'Contraseña';
const _submit = 'Entrar';
const _rejected = 'Correo o contraseña incorrectos.';
const _missing = 'Escribe este dato.';

String _retryIn(String time) => 'Podrás intentar de nuevo en $time.';

void main() {
  late AuthHarness harness;

  setUp(() => harness = AuthHarness());

  Finder primary() => find.widgetWithText(FilledButton, _submit);

  Future<void> tapPrimary(WidgetTester tester) async {
    await tester.ensureVisible(primary());
    await tester.tap(primary());
    await tester.pumpAndSettle();
  }

  Future<void> fill(WidgetTester tester) async {
    await tester.enterText(inputLabeled(_email), ' ana@example.mx ');
    await tester.enterText(inputLabeled(_password), syntheticPassword());
  }

  TextField input(WidgetTester tester, String label) =>
      tester.widget<TextField>(inputLabeled(label));

  group('layout', () {
    testWidgets('signed out, the app opens on the login form', (tester) async {
      await harness.pump(tester);

      expect(find.byType(LoginScreen), findsOneWidget);
      expect(find.text(_title), findsOneWidget);
      final tops = [
        for (final label in [_email, _password])
          tester.getTopLeft(find.text(label)).dy,
      ];
      expect(tops.first, lessThan(tops.last), reason: 'email first');
    });

    testWidgets('fields carry autofill hints and the right keyboards', (
      tester,
    ) async {
      await harness.pump(tester);

      final email = input(tester, _email);
      expect(email.autofillHints, [AutofillHints.email]);
      expect(email.keyboardType, TextInputType.emailAddress);
      final password = input(tester, _password);
      expect(password.autofillHints, [AutofillHints.password]);
      expect(password.obscureText, isTrue);
    });

    testWidgets('the password can be shown and hidden', (tester) async {
      await harness.pump(tester);

      await tester.tap(find.byTooltip('Mostrar contraseña'));
      await tester.pump();
      expect(input(tester, _password).obscureText, isFalse);

      await tester.tap(find.byTooltip('Ocultar contraseña'));
      await tester.pump();
      expect(input(tester, _password).obscureText, isTrue);
    });

    testWidgets('"Crear cuenta" opens the sign-up form', (tester) async {
      await harness.pump(tester);

      await tester.tap(find.widgetWithText(TextButton, 'Crear cuenta'));
      await tester.pumpAndSettle();

      expect(find.text('Acepto el aviso de privacidad'), findsOneWidget);
    });
  });

  group('submit', () {
    testWidgets('valid credentials sign in and land on home', (tester) async {
      await harness.pump(tester);
      await fill(tester);

      await tapPrimary(tester);

      final login = harness.repository.logins.single;
      expect(login.email, 'ana@example.mx', reason: 'trimmed');
      expect(login.password, syntheticPassword(), reason: 'verbatim');
      expect(harness.tokens.token, 'refresh', reason: 'secure store');
      expect(harness.session.accessToken, 'access', reason: 'memory');
      expect(find.byType(HomeScreen), findsOneWidget);
    });

    testWidgets('Enter in the password field submits', (tester) async {
      await harness.pump(tester);
      await fill(tester);

      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();

      expect(harness.repository.logins, hasLength(1));
    });

    testWidgets('empty fields ask for both and send nothing', (tester) async {
      await harness.pump(tester);

      await tapPrimary(tester);

      expect(find.text('$_email: $_missing'), findsOneWidget);
      expect(find.text('$_password: $_missing'), findsOneWidget);
      expect(harness.repository.logins, isEmpty);
    });

    testWidgets('a malformed email shows the format rule on blur', (
      tester,
    ) async {
      await harness.pump(tester);

      await tester.enterText(inputLabeled(_email), 'ana@');
      await tester.testTextInput.receiveAction(TextInputAction.next);
      await tester.pumpAndSettle();

      expect(find.text('Escribe un correo como nombre@ejemplo.mx.'), findsOne);
    });

    testWidgets('bad credentials show one generic message; nothing kept', (
      tester,
    ) async {
      harness.repository.loginError = const ApiUnauthorized();
      await harness.pump(tester);
      await fill(tester);

      await tapPrimary(tester);

      expect(find.text(_rejected), findsOneWidget);
      expect(find.byType(LoginScreen), findsOneWidget);
      expect(harness.tokens.token, isNull);
      expect(harness.session.status, SessionStatus.signedOut);
    });

    testWidgets('too many tries say how long to wait', (tester) async {
      harness.repository.loginError = const ApiRateLimited(
        Duration(minutes: 3),
      );
      await harness.pump(tester);
      await fill(tester);

      await tapPrimary(tester);

      expect(
        find.text('Demasiados intentos. Espera 3 min y vuelve a intentarlo.'),
        findsOneWidget,
      );
    });

    // M1-15 QA regression: on the emulator the login 429 showed a static
    // "Espera 3 min" that never moved, and every tap during the wait went to
    // the server (spending the per-network login budget of everyone behind
    // the same network). The wait now counts down like the delete dialog's.
    testWidgets('too many tries count the Retry-After down; a tap during '
        'the wait sends nothing and says how long is left', (tester) async {
      harness.repository.loginError = const ApiRateLimited(
        Duration(minutes: 3),
      );
      await harness.pump(tester);
      await fill(tester);
      await tapPrimary(tester);

      expect(find.text(_retryIn('3:00')), findsOneWidget);
      await harness.elapse(tester, const Duration(minutes: 1));
      expect(find.text(_retryIn('2:00')), findsOneWidget);

      tester.takeAnnouncements();
      await tapPrimary(tester);
      expect(harness.repository.logins, hasLength(1), reason: 'not sent');
      expect(
        tester.takeAnnouncements().map((a) => a.message),
        contains(_retryIn('2:00')),
      );

      await harness.elapse(tester, const Duration(minutes: 2));
      expect(find.textContaining('Podrás intentar'), findsNothing);

      harness.repository.loginError = null;
      await tapPrimary(tester);
      expect(harness.repository.logins, hasLength(2));
      expect(find.byType(HomeScreen), findsOneWidget);
    });

    testWidgets('no connection says so', (tester) async {
      harness.repository.loginError = const ApiUnreachable();
      await harness.pump(tester);
      await fill(tester);

      await tapPrimary(tester);

      expect(
        find.text(
          'No hay conexión con el servidor. Revisa la conexión a internet '
          'e intenta de nuevo.',
        ),
        findsOneWidget,
      );
    });

    testWidgets('the primary stays enabled and ignores taps while busy', (
      tester,
    ) async {
      harness.repository.loginGate = Completer<void>();
      await harness.pump(tester);
      await fill(tester);

      await tester.tap(primary());
      await tester.pump();
      await tester.tap(primary(), warnIfMissed: false);
      await tester.pump();

      expect(harness.repository.logins, hasLength(1));
      harness.repository.loginGate!.complete();
      await tester.pumpAndSettle();
    });
  });

  group('accessibility', () {
    for (final brightness in Brightness.values) {
      testWidgets('meets the tap-target, label and contrast guidelines '
          '(${brightness.name})', (tester) async {
        final semantics = tester.ensureSemantics();
        await harness.pump(tester, brightness: brightness);
        await tapPrimary(tester);

        await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
        await expectLater(tester, meetsGuideline(iOSTapTargetGuideline));
        await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
        await expectLater(tester, meetsGuideline(textContrastGuideline));
        semantics.dispose();
      });
    }

    testWidgets('survives 200 % text', (tester) async {
      tester.platformDispatcher.textScaleFactorTestValue = 2;
      addTearDown(tester.platformDispatcher.clearTextScaleFactorTestValue);

      await harness.pump(tester);
      await tapPrimary(tester);

      expect(tester.takeException(), isNull);
    });
  });

  testWidgets('module off: the login location is not found', (tester) async {
    await harness.pump(tester, location: AuthRoutes.login, enabled: false);

    expect(find.text('Página no encontrada'), findsOneWidget);
  });
}
