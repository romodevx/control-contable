import 'dart:async';

import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/features/auth/application/pending_sign_up.dart';
import 'package:control_contable/features/auth/auth_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/synthetic_password.dart';
import 'pump_auth.dart';

const _email = 'ana@example.mx';
const _code = 'Código de verificación';
const _confirm = 'Confirmar';
const _password = 'Contraseña';
const _policy =
    'Esta contraseña es muy común o apareció en una filtración. Elige otra.';
const _resend = 'Enviar otro código';
const _rejected =
    'El código no es válido o ya venció. Revisa el correo o pide uno nuevo.';

void main() {
  late AuthHarness harness;

  setUp(() => harness = AuthHarness());

  /// The draft a successful register leaves behind.
  void startDraft() => harness.container
      .read(pendingSignUpProvider.notifier)
      .start(
        SignUpDraft(email: _email, password: syntheticPassword(), name: 'Ana'),
      );

  /// Opens the code screen the way a successful register does.
  Future<void> open(WidgetTester tester, {bool pending = true}) async {
    await harness.pump(tester);
    if (pending) startDraft();
    harness.container.read(appRouterProvider).go(AuthRoutes.verifyEmail);
    await tester.pumpAndSettle();
  }

  Finder code() => inputLabeled(_code);

  /// Leaves the app (e.g. to read the code in the mail app) and comes back.
  Future<void> backgroundAndReturn(WidgetTester tester) async {
    [
      AppLifecycleState.inactive,
      AppLifecycleState.hidden,
      AppLifecycleState.paused,
      AppLifecycleState.hidden,
      AppLifecycleState.inactive,
      AppLifecycleState.resumed,
    ].forEach(tester.binding.handleAppLifecycleStateChanged);
    await tester.pump();
  }

  testWidgets('says where the code went and how long it lasts', (tester) async {
    await open(tester);

    expect(
      find.text(
        'Enviamos un código de 6 dígitos a $_email. Vence en 15 minutos.',
      ),
      findsOneWidget,
    );
  });

  testWidgets('the code field asks for one-time-code autofill and digits', (
    tester,
  ) async {
    await open(tester);

    final field = tester.widget<TextField>(code());
    expect(field.autofillHints, [AutofillHints.oneTimeCode]);
    expect(field.keyboardType, TextInputType.number);
    expect(
      field.inputFormatters,
      contains(isA<LengthLimitingTextInputFormatter>()),
    );

    await tester.enterText(code(), '12a');
    expect(find.text('12'), findsOneWidget, reason: 'digits only');
  });

  testWidgets('six digits submit by themselves and sign in', (tester) async {
    await open(tester);

    await tester.enterText(code(), '482913');
    await tester.pumpAndSettle();

    final request = harness.repository.verified.single;
    expect((request.email, request.code), (_email, '482913'));
    expect(request.password, syntheticPassword(), reason: 'D8: set here');
    expect(request.name, 'Ana');
    expect(harness.sessions.saved, hasLength(1));
    expect(harness.container.read(pendingSignUpProvider), isNull);
    expect(find.text('Inicio'), findsOneWidget, reason: 'home');
  });

  testWidgets('Confirmar with no code asks for it and sends nothing', (
    tester,
  ) async {
    await open(tester);

    await tester.tap(find.widgetWithText(FilledButton, _confirm));
    await tester.pumpAndSettle();

    expect(find.text('$_code: Escribe este dato.'), findsOneWidget);
    expect(harness.repository.verified, isEmpty);
  });

  testWidgets('a short code shows the format rule on blur', (tester) async {
    await open(tester);

    await tester.enterText(code(), '123');
    await tester.testTextInput.receiveAction(TextInputAction.done);
    await tester.pumpAndSettle();

    expect(find.text('Escribe los 6 dígitos del código.'), findsWidgets);
    expect(harness.repository.verified, isEmpty);
  });

  testWidgets('a wrong code shows one generic message on the field', (
    tester,
  ) async {
    harness.repository.verifyError = const ApiUnauthorized();
    await open(tester);

    await tester.enterText(code(), '000000');
    await tester.pumpAndSettle();

    expect(find.text(_rejected), findsOneWidget, reason: 'inline');
    expect(find.text('$_code: $_rejected'), findsOneWidget);
    expect(harness.sessions.saved, isEmpty);
  });

  group('password chosen at sign-up (D8)', () {
    testWidgets('a rejected password reveals its field; the code is kept', (
      tester,
    ) async {
      harness.repository.verifyError = const ApiValidationFailure([
        api.FieldError(
          field: 'password',
          code: 'password_policy',
          message: 'Use at least 12 characters.',
        ),
      ]);
      await open(tester);
      expect(find.text(_password), findsNothing, reason: 'hidden at first');

      await tester.enterText(code(), '482913');
      await tester.pumpAndSettle();

      expect(find.text('$_password: $_policy'), findsOneWidget);
      expect(find.text(_policy), findsOneWidget, reason: 'inline');
      expect(find.text(_rejected), findsNothing, reason: 'not a code error');
      expect(find.text('482913'), findsOneWidget);

      harness.repository.verifyError = null;
      await tester.enterText(inputLabeled(_password), syntheticPassword(20));
      await tester.tap(find.widgetWithText(FilledButton, _confirm));
      await tester.pumpAndSettle();

      expect(harness.repository.verified.last.password, syntheticPassword(20));
      expect(harness.sessions.saved, hasLength(1));
    });

    testWidgets('a 400 on the code stays on the code field', (tester) async {
      harness.repository.verifyError = const ApiValidationFailure([
        api.FieldError(field: 'code', code: 'pattern', message: 'Bad code.'),
      ]);
      await open(tester);

      await tester.enterText(code(), '482913');
      await tester.pumpAndSettle();

      expect(find.text('$_code: Escribe los 6 dígitos del código.'), findsOne);
      expect(find.text(_password), findsNothing);
    });

    testWidgets('going to the background forgets the password', (tester) async {
      await open(tester);

      await backgroundAndReturn(tester);
      await tester.pumpAndSettle();

      expect(harness.container.read(pendingSignUpProvider)!.password, isNull);
      expect(harness.container.read(pendingSignUpProvider)!.email, _email);
      expect(find.text(_password), findsOneWidget, reason: 'asked again');

      await tester.enterText(code(), '482913');
      await tester.pumpAndSettle();
      expect(harness.repository.verified, isEmpty, reason: 'waits for it');

      await tester.enterText(inputLabeled(_password), syntheticPassword());
      await tester.tap(find.widgetWithText(FilledButton, _confirm));
      await tester.pumpAndSettle();
      expect(harness.repository.verified.single.password, syntheticPassword());
    });
  });

  group('resend', () {
    testWidgets('waits 60 s, then sends a new code and waits again', (
      tester,
    ) async {
      await open(tester);

      expect(find.text('Otro código disponible en 1:00.'), findsOneWidget);
      expect(find.widgetWithText(TextButton, _resend), findsNothing);

      await harness.elapse(tester, const Duration(seconds: 15));
      expect(find.text('Otro código disponible en 0:45.'), findsOneWidget);

      await harness.elapse(tester, const Duration(seconds: 45));
      await tester.tap(find.widgetWithText(TextButton, _resend));
      await tester.pumpAndSettle();

      expect(harness.repository.resent, [_email]);
      expect(
        find.text(
          'Si hay un registro pendiente con ese correo, enviamos un código '
          'nuevo.',
        ),
        findsOneWidget,
      );
      expect(find.text('Otro código disponible en 1:00.'), findsOneWidget);
      await harness.elapse(tester, const Duration(seconds: 60));
    });

    testWidgets('a rate-limited resend waits for Retry-After', (tester) async {
      harness.repository.resendError = const ApiRateLimited(
        Duration(seconds: 120),
      );
      await open(tester);
      await harness.elapse(tester, const Duration(seconds: 60));

      await tester.tap(find.widgetWithText(TextButton, _resend));
      await tester.pumpAndSettle();

      expect(find.text('Otro código disponible en 2:00.'), findsOneWidget);
      await harness.elapse(tester, const Duration(seconds: 120));
    });
  });

  testWidgets('"Cambiar correo" goes back to the sign-up form', (tester) async {
    await harness.pump(tester, location: AuthRoutes.register);
    startDraft();
    unawaited(
      harness.container
          .read(appRouterProvider)
          .push<void>(AuthRoutes.verifyEmail),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.widgetWithText(TextButton, 'Cambiar correo'));
    await tester.pumpAndSettle();

    expect(find.text('Correo electrónico'), findsOneWidget);
  });

  testWidgets('without a sign-up in progress it points back to register', (
    tester,
  ) async {
    await open(tester, pending: false);

    expect(find.text('No hay un registro en curso.'), findsOneWidget);
    await tester.tap(find.widgetWithText(FilledButton, 'Crear cuenta'));
    await tester.pumpAndSettle();

    expect(find.text('Correo electrónico'), findsOneWidget);
  });

  for (final brightness in Brightness.values) {
    testWidgets('meets the tap-target, label and contrast guidelines '
        '(${brightness.name})', (tester) async {
      final semantics = tester.ensureSemantics();
      await harness.pump(tester, brightness: brightness);
      startDraft();
      harness.container.read(appRouterProvider).go(AuthRoutes.verifyEmail);
      await tester.pumpAndSettle();
      await harness.elapse(tester, const Duration(seconds: 60));

      await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
      await expectLater(tester, meetsGuideline(iOSTapTargetGuideline));
      await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
      await expectLater(tester, meetsGuideline(textContrastGuideline));
      semantics.dispose();
    });
  }
}
