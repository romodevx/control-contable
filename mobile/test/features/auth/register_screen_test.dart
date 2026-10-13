import 'dart:async';

import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/api/generated/models/legal_document_kind.dart';
import 'package:control_contable/core/forms/app_checkbox_field.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/auth/application/pending_sign_up.dart';
import 'package:control_contable/features/auth/auth_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/synthetic_password.dart';
import 'pump_auth.dart';

const _name = 'Nombre (opcional)';
const _email = 'Correo electrónico';
const _password = 'Contraseña';
const _privacy = 'Acepto el aviso de privacidad';
const _terms = 'Acepto los términos y condiciones';
const _readPrivacy = 'Leer el aviso de privacidad';
const _readTerms = 'Leer los términos y condiciones';
const _submit = 'Crear cuenta';
const _missing = 'Escribe este dato.';
const _notAccepted = 'Marca esta casilla para continuar.';

void main() {
  late AuthHarness harness;

  setUp(() => harness = AuthHarness());

  Future<void> open(WidgetTester tester) =>
      harness.pump(tester, location: AuthRoutes.register);

  Finder primary() => find.widgetWithText(FilledButton, _submit);

  Future<void> tapPrimary(WidgetTester tester) async {
    await tester.ensureVisible(primary());
    await tester.tap(primary());
    await tester.pumpAndSettle();
  }

  Future<void> check(WidgetTester tester, String label) async {
    await tester.ensureVisible(find.text(label));
    await tester.tap(find.text(label));
    await tester.pump();
  }

  Future<void> fillValid(WidgetTester tester, {String name = ''}) async {
    await tester.enterText(inputLabeled(_name), name);
    await tester.enterText(inputLabeled(_email), '  ana@example.mx ');
    await tester.enterText(inputLabeled(_password), '${syntheticPassword()} ');
    await check(tester, _privacy);
    await check(tester, _terms);
  }

  bool isChecked(WidgetTester tester, String label) => tester
      .widget<Checkbox>(
        find.descendant(
          of: find.ancestor(
            of: find.text(label),
            matching: find.byType(AppCheckboxField),
          ),
          matching: find.byType(Checkbox),
        ),
      )
      .value!;

  group('layout', () {
    testWidgets('name, email and password come before the consents', (
      tester,
    ) async {
      await open(tester);

      final tops = [
        for (final label in [_name, _email, _password, _privacy, _terms])
          tester.getTopLeft(find.text(label)).dy,
      ];
      expect(tops, orderedEquals([...tops]..sort()));
      expect(find.widgetWithText(TextButton, _readPrivacy), findsOneWidget);
      expect(find.widgetWithText(TextButton, _readTerms), findsOneWidget);
    });

    testWidgets('email and password ask for the right keyboard and autofill', (
      tester,
    ) async {
      await open(tester);

      final email = tester.widget<TextField>(inputLabeled(_email));
      expect(email.keyboardType, TextInputType.emailAddress);
      expect(email.autofillHints, [AutofillHints.email]);
      final password = tester.widget<TextField>(inputLabeled(_password));
      expect(password.obscureText, isTrue);
      expect(password.autofillHints, [AutofillHints.newPassword]);
      expect(find.textContaining('Al menos 12 caracteres'), findsOneWidget);
    });
  });

  group('client checks', () {
    testWidgets('an empty submit lists every missing item and sends nothing', (
      tester,
    ) async {
      await open(tester);

      await tapPrimary(tester);

      expect(find.text('Revisa 4 datos antes de guardar'), findsOneWidget);
      for (final item in [
        '$_email: $_missing',
        '$_password: $_missing',
        '$_privacy: $_notAccepted',
        '$_terms: $_notAccepted',
      ]) {
        expect(find.text(item), findsOneWidget, reason: item);
      }
      expect(harness.repository.registered, isEmpty);
    });

    testWidgets('a malformed email shows its error on leaving the field', (
      tester,
    ) async {
      await open(tester);

      await tester.enterText(inputLabeled(_email), 'ana@');
      await tester.pump();
      expect(find.textContaining('nombre@ejemplo.mx'), findsNothing);

      await tester.tap(inputLabeled(_password));
      await tester.pump();
      expect(
        find.text('Escribe un correo como nombre@ejemplo.mx.'),
        findsOneWidget,
      );
    });

    testWidgets('a short password shows the length rule on blur', (
      tester,
    ) async {
      await open(tester);

      await tester.enterText(inputLabeled(_password), syntheticPassword(5));
      await tester.tap(inputLabeled(_email));
      await tester.pump();

      expect(find.text('Usa al menos 12 caracteres.'), findsOneWidget);
    });
  });

  group('a complete submit', () {
    testWidgets('sends the versions shown and opens the code screen', (
      tester,
    ) async {
      await open(tester);
      await fillValid(tester);

      await tapPrimary(tester);

      final request = harness.repository.registered.single;
      expect(request.email, 'ana@example.mx');
      expect(
        request.password,
        '${syntheticPassword()} ',
        reason: 'never trimmed',
      );
      expect(request.name, isNull);
      expect(request.acceptedDocuments.map((d) => (d.kind, d.version)), [
        (LegalDocumentKind.privacyNoticeIntegral, '2026-10-01'),
        (LegalDocumentKind.terms, '2026-10-01'),
      ]);
      expect(find.text('Confirmar correo'), findsOneWidget);
      expect(harness.container.read(pendingSignUpProvider), 'ana@example.mx');
    });

    testWidgets('sends the name trimmed when given', (tester) async {
      await open(tester);
      await fillValid(tester, name: '  Ana Gómez ');

      await tapPrimary(tester);

      expect(harness.repository.registered.single.name, 'Ana Gómez');
    });

    testWidgets('shows busy and ignores repeat taps while sending', (
      tester,
    ) async {
      final gate = harness.repository.registerGate = Completer<void>();
      await open(tester);
      await fillValid(tester);

      await tester.ensureVisible(primary());
      await tester.tap(primary());
      await tester.pump();
      await tester.tap(primary());
      await tester.pump();

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(tester.widget<FilledButton>(primary()).onPressed, isNotNull);
      expect(harness.repository.registered, hasLength(1));
      gate.complete();
      await tester.pumpAndSettle();
    });
  });

  group('server verdicts', () {
    testWidgets('a weak password shows on its field, in Spanish', (
      tester,
    ) async {
      harness.repository.registerError = const ApiValidationFailure([
        api.FieldError(
          field: 'password',
          code: 'password_policy',
          message: 'Use at least 12 characters.',
        ),
      ]);
      await open(tester);
      await fillValid(tester);

      await tapPrimary(tester);

      const policy =
          'Esta contraseña es muy común o apareció en una filtración. '
          'Elige otra.';
      expect(find.text('$_password: $policy'), findsOneWidget);
      expect(find.text(policy), findsOneWidget, reason: 'inline');
      expect(find.textContaining('Use at least'), findsNothing);
      expect(find.text('Confirmar correo'), findsNothing);
    });

    testWidgets('stale documents are reloaded and must be accepted again', (
      tester,
    ) async {
      harness.repository.registerError = const ApiValidationFailure([
        api.FieldError(
          field: 'accepted_documents.1.version',
          code: 'document_version_not_current',
          message: 'Accept the current privacy notice and terms.',
        ),
      ]);
      await open(tester);
      await fillValid(tester);

      await tapPrimary(tester);

      expect(harness.repository.documentCalls, 2);
      expect(isChecked(tester, _privacy), isFalse);
      expect(isChecked(tester, _terms), isFalse);
      expect(
        find.text(
          'Los documentos se actualizaron. Revísalos y acéptalos de nuevo.',
        ),
        findsWidgets,
      );
    });

    testWidgets('too many tries says how long to wait', (tester) async {
      harness.repository.registerError = const ApiRateLimited(
        Duration(minutes: 5),
      );
      await open(tester);
      await fillValid(tester);

      await tapPrimary(tester);

      expect(
        find.text('Demasiados intentos. Espera 5 min y vuelve a intentarlo.'),
        findsOneWidget,
      );
    });

    testWidgets('no connection keeps what was typed', (tester) async {
      harness.repository.registerError = const ApiUnreachable();
      await open(tester);
      await fillValid(tester);

      await tapPrimary(tester);

      expect(
        find.text(
          'No hay conexión con el servidor. Revisa la conexión a internet '
          'e intenta de nuevo.',
        ),
        findsOneWidget,
      );
      expect(find.text('ana@example.mx'), findsOneWidget);
    });
  });

  group('legal documents', () {
    testWidgets('a failed load offers a retry that loads them', (tester) async {
      harness.repository.documentsError = const ApiUnreachable();
      await open(tester);

      expect(
        find.text('No se pudieron cargar los documentos legales.'),
        findsOneWidget,
      );
      expect(find.widgetWithText(TextButton, _readTerms), findsNothing);

      harness.repository.documentsError = null;
      await tester.ensureVisible(find.text('Reintentar'));
      await tester.tap(find.text('Reintentar'));
      await tester.pumpAndSettle();

      expect(find.widgetWithText(TextButton, _readTerms), findsOneWidget);
    });

    testWidgets('a submit before they load says so and sends nothing', (
      tester,
    ) async {
      harness.repository.documentsError = const ApiUnreachable();
      await open(tester);
      await fillValid(tester);

      await tapPrimary(tester);

      expect(harness.repository.registered, isEmpty);
      expect(
        find.text(
          'Los documentos legales no se han cargado. Reintenta para continuar.',
        ),
        findsOneWidget,
      );
    });

    testWidgets('each link opens its document', (tester) async {
      await open(tester);

      await tester.ensureVisible(find.text(_readTerms));
      await tester.tap(find.text(_readTerms));
      await tester.pump();

      expect(harness.opened, [Uri.parse('https://example.mx/legal/terminos')]);
    });

    testWidgets('a link that cannot open says so', (tester) async {
      harness.canOpen = false;
      await open(tester);

      await tester.ensureVisible(find.text(_readPrivacy));
      await tester.tap(find.text(_readPrivacy));
      await tester.pumpAndSettle();

      expect(
        find.text('No se pudo abrir el documento. Intenta de nuevo.'),
        findsOneWidget,
      );
    });
  });

  group('accessibility', () {
    for (final brightness in Brightness.values) {
      testWidgets('meets the tap-target, label and contrast guidelines '
          '(${brightness.name})', (tester) async {
        final semantics = tester.ensureSemantics();
        await harness.pump(
          tester,
          location: AuthRoutes.register,
          brightness: brightness,
        );
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

      await open(tester);
      await tapPrimary(tester);

      expect(tester.takeException(), isNull);
    });
  });
}
