import 'dart:async';

import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/auth/auth_routes.dart';
import 'package:control_contable/features/auth/presentation/profile_screen.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'pump_auth.dart';

const _entry = 'Mi cuenta';
const _name = 'Nombre';
const _save = 'Guardar cambios';
const _saved = 'Guardamos tu nombre.';
const _missing = 'Escribe este dato.';
const _tooLong = 'Escribe un nombre de hasta 120 caracteres.';
const _loadFailed = 'No se pudo cargar tu cuenta.';
const _privacy = 'Leer el aviso de privacidad';
const _terms = 'Leer los términos y condiciones';
const _deleteEntry = 'Eliminar mi cuenta';

/// Smallest tap target the house rules allow (WCAG 2.2 asks 24; we use 44).
const double _minTarget = 44;

void main() {
  late AuthHarness harness;

  setUp(() => harness = AuthHarness());

  Future<void> openProfile(WidgetTester tester) =>
      harness.pump(tester, signedIn: true, location: AuthRoutes.profile);

  Future<void> tapText(WidgetTester tester, String text) async {
    await tester.ensureVisible(find.text(text));
    await tester.tap(find.text(text));
    await tester.pumpAndSettle();
  }

  group('flag', () {
    testWidgets('off: the profile is not found and never loads', (
      tester,
    ) async {
      await harness.pump(
        tester,
        signedIn: true,
        enabled: false,
        location: AuthRoutes.profile,
      );

      expect(find.text('Página no encontrada'), findsOneWidget);
      expect(harness.profile.calls, isEmpty);
    });

    testWidgets('off: home has no way into the account', (tester) async {
      await harness.pump(tester, enabled: false);

      expect(find.byTooltip(_entry), findsNothing);
    });

    testWidgets('on: home opens the account screen', (tester) async {
      await harness.pump(tester, signedIn: true);

      await tester.tap(find.byTooltip(_entry));
      await tester.pumpAndSettle();

      expect(find.byType(ProfileScreen), findsOneWidget);
      expect(find.text(_entry), findsOneWidget, reason: 'the title');
    });
  });

  testWidgets('signed out, the profile asks to sign in first', (tester) async {
    await harness.pump(tester, location: AuthRoutes.profile);

    expect(find.text('Iniciar sesión'), findsWidgets);
    expect(harness.profile.calls, isEmpty);
  });

  group('loading', () {
    testWidgets('shows a status while the profile loads', (tester) async {
      harness.profile.getMeGate = Completer<void>();
      await harness.pump(tester, signedIn: true);
      unawaited(harness.router.push(AuthRoutes.profile));
      await tester.pump();
      await tester.pump();

      expect(find.bySemanticsLabel('Cargando tu cuenta…'), findsOneWidget);

      harness.profile.getMeGate!.complete();
      await tester.pumpAndSettle();
      expect(inputLabeled(_name), findsOneWidget);
    });

    testWidgets('a failed load offers a retry that loads again', (
      tester,
    ) async {
      harness.profile.getMeError = const ApiUnreachable();
      await openProfile(tester);

      expect(find.text(_loadFailed), findsOneWidget);

      harness.profile.getMeError = null;
      await tapText(tester, 'Reintentar');

      expect(find.text(_loadFailed), findsNothing);
      expect(inputLabeled(_name), findsOneWidget);
    });
  });

  group('profile', () {
    testWidgets('the name is editable, the email is read-only', (tester) async {
      await openProfile(tester);

      final name = tester.widget<TextField>(inputLabeled(_name));
      expect(name.controller!.text, 'Ana Gómez');
      expect(find.text('ana@example.mx'), findsOneWidget);
      expect(
        find.widgetWithText(TextField, 'ana@example.mx'),
        findsNothing,
        reason: 'the email is text, not an input',
      );
      expect(find.text('El correo no se puede cambiar.'), findsOneWidget);
    });

    testWidgets('saving a new name patches it and confirms', (tester) async {
      await openProfile(tester);

      await tester.enterText(inputLabeled(_name), '  Ana Gómez Ruiz ');
      await tapText(tester, _save);

      expect(harness.profile.names, ['Ana Gómez Ruiz']);
      expect(find.text(_saved), findsOneWidget);
    });

    testWidgets('an empty name asks for it and sends nothing', (tester) async {
      await openProfile(tester);

      await tester.enterText(inputLabeled(_name), '');
      await tapText(tester, _save);

      expect(find.text(_missing), findsWidgets);
      expect(harness.profile.names, isEmpty);
    });

    testWidgets('the server verdict on the name lands on the field', (
      tester,
    ) async {
      harness.profile.updateError = const ApiValidationFailure([
        api.FieldError(field: 'name', code: 'maxLength', message: 'Too long'),
      ]);
      await openProfile(tester);

      await tester.enterText(inputLabeled(_name), 'Ana');
      await tapText(tester, _save);

      expect(find.text(_tooLong), findsWidgets);
      expect(find.text(_saved), findsNothing);
    });

    testWidgets('an unreachable server says so', (tester) async {
      harness.profile.updateError = const ApiUnreachable();
      await openProfile(tester);

      await tapText(tester, _save);

      expect(find.textContaining('No hay conexión'), findsOneWidget);
    });
  });

  group('legal documents', () {
    testWidgets('each link opens its current document', (tester) async {
      await openProfile(tester);

      await tapText(tester, _privacy);
      await tapText(tester, _terms);

      expect(harness.opened.map((uri) => uri.path), [
        '/legal/aviso-integral',
        '/legal/terminos',
      ]);
    });

    testWidgets('a failed load offers a retry', (tester) async {
      harness.repository.documentsError = const ApiUnreachable();
      await openProfile(tester);

      expect(
        find.text('No se pudieron cargar los documentos legales.'),
        findsOneWidget,
      );
      harness.repository.documentsError = null;
      await tester.ensureVisible(find.text('Reintentar'));
      await tester.tap(find.text('Reintentar'));
      await tester.pumpAndSettle();

      expect(find.text(_privacy), findsOneWidget);
    });
  });

  testWidgets('meets the tap-target, labeling and contrast guidelines', (
    tester,
  ) async {
    final semantics = tester.ensureSemantics();
    await openProfile(tester);

    for (final label in [_save, _privacy, _terms, _deleteEntry]) {
      final size = tester.getSize(
        find.ancestor(
          of: find.text(label),
          matching: find.bySubtype<ButtonStyleButton>(),
        ),
      );
      expect(size.height, greaterThanOrEqualTo(_minTarget), reason: label);
    }
    await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
    await expectLater(tester, meetsGuideline(textContrastGuideline));
    semantics.dispose();
  });

  testWidgets('renders in dark mode and at 200 % text without overflow', (
    tester,
  ) async {
    tester.platformDispatcher.textScaleFactorTestValue = 2;
    addTearDown(tester.platformDispatcher.clearTextScaleFactorTestValue);

    await harness.pump(
      tester,
      signedIn: true,
      location: AuthRoutes.profile,
      brightness: Brightness.dark,
    );

    expect(tester.takeException(), isNull);
    expect(find.byType(ProfileScreen), findsOneWidget);
  });
}
