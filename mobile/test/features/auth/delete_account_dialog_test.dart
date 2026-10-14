import 'dart:async';

import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/auth/auth_routes.dart';
import 'package:control_contable/features/auth/presentation/goodbye_screen.dart';
import 'package:control_contable/features/auth/presentation/login_screen.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/synthetic_password.dart';
import 'pump_auth.dart';

const _deleteEntry = 'Eliminar mi cuenta';
const _title = '¿Eliminar tu cuenta?';
const _password = 'Contraseña actual';
const _confirm = 'Eliminar cuenta';
const _cancel = 'Cancelar';
const _wrongPassword = 'La contraseña no es correcta.';
const _missing = 'Escribe este dato.';

void main() {
  late AuthHarness harness;

  setUp(() => harness = AuthHarness());

  Finder confirmButton() => find.ancestor(
    of: find.text(_confirm),
    matching: find.byType(FilledButton),
  );

  Future<void> openDialog(WidgetTester tester) async {
    await harness.pump(tester, signedIn: true, location: AuthRoutes.profile);
    await tester.ensureVisible(find.text(_deleteEntry));
    await tester.tap(find.text(_deleteEntry));
    await tester.pumpAndSettle();
  }

  Future<void> confirm(WidgetTester tester) async {
    await tester.tap(confirmButton());
    await tester.pumpAndSettle();
  }

  group('the dialog', () {
    testWidgets('is a modal labeled by its title that explains the loss', (
      tester,
    ) async {
      await openDialog(tester);

      expect(find.byType(AlertDialog), findsOneWidget);
      expect(find.text(_title), findsOneWidget);
      expect(find.textContaining('No se puede deshacer'), findsOneWidget);
      expect(
        find.textContaining('La prueba gratuita no se reinicia'),
        findsOneWidget,
      );
      expect(find.textContaining('suscripción'), findsOneWidget);
    });

    testWidgets('focuses the password field first', (tester) async {
      await openDialog(tester);

      final field = tester.widget<TextField>(inputLabeled(_password));
      expect(field.focusNode!.hasFocus, isTrue);
      expect(field.obscureText, isTrue);
      expect(field.autofillHints, contains(AutofillHints.password));
    });

    testWidgets('the primary is destructive and never disabled', (
      tester,
    ) async {
      await openDialog(tester);

      final button = tester.widget<FilledButton>(confirmButton());
      final context = tester.element(confirmButton());
      expect(button.onPressed, isNotNull);
      expect(
        button.style?.backgroundColor?.resolve({}),
        context.tokens.colors.error,
      );
      expect(tester.getSize(confirmButton()).height, greaterThanOrEqualTo(44));
    });

    testWidgets('Cancelar closes it and focus returns to the trigger', (
      tester,
    ) async {
      await openDialog(tester);

      await tester.tap(find.text(_cancel).last);
      await tester.pumpAndSettle();

      expect(find.byType(AlertDialog), findsNothing);
      final trigger = find.ancestor(
        of: find.text(_deleteEntry),
        matching: find.byType(Focus),
      );
      expect(
        tester
            .widgetList<Focus>(trigger)
            .any((f) => f.focusNode?.hasFocus ?? false),
        isTrue,
      );
      expect(harness.profile.deletePasswords, isEmpty);
    });

    testWidgets('Esc closes it', (tester) async {
      await openDialog(tester);

      await tester.sendKeyEvent(LogicalKeyboardKey.escape);
      await tester.pumpAndSettle();

      expect(find.byType(AlertDialog), findsNothing);
    });
  });

  group('confirming', () {
    testWidgets('an empty password asks for it and sends nothing', (
      tester,
    ) async {
      await openDialog(tester);

      await confirm(tester);

      expect(find.text(_missing), findsWidgets);
      expect(harness.profile.deletePasswords, isEmpty);
    });

    testWidgets('a wrong password: one try, field cleared and refocused, '
        'still signed in', (tester) async {
      harness.profile.deleteError = const ApiUnauthorized();
      await openDialog(tester);

      await tester.enterText(inputLabeled(_password), syntheticPassword());
      await confirm(tester);

      expect(harness.profile.deletePasswords, hasLength(1));
      expect(find.text(_wrongPassword), findsWidgets);
      final field = tester.widget<TextField>(inputLabeled(_password));
      expect(field.controller!.text, isEmpty);
      expect(field.focusNode!.hasFocus, isTrue);
      expect(harness.session.status, SessionStatus.signedIn);
      expect(find.byType(AlertDialog), findsOneWidget);
    });

    testWidgets('too many tries say how long to wait', (tester) async {
      harness.profile.deleteError = const ApiRateLimited(Duration(minutes: 5));
      await openDialog(tester);

      await tester.enterText(inputLabeled(_password), syntheticPassword());
      await confirm(tester);

      expect(find.textContaining('Espera 5 min'), findsOneWidget);
    });

    testWidgets('while deleting it shows busy and ignores repeat taps', (
      tester,
    ) async {
      harness.profile.deleteGate = Completer<void>();
      await openDialog(tester);
      await tester.enterText(inputLabeled(_password), syntheticPassword());

      await tester.tap(confirmButton());
      await tester.pump();
      await tester.tap(confirmButton());
      await tester.pump();

      expect(find.bySemanticsLabel('Eliminando cuenta…'), findsOneWidget);
      expect(tester.widget<FilledButton>(confirmButton()).onPressed, isNotNull);
      expect(harness.profile.deletePasswords, hasLength(1));
      harness.profile.deleteGate!.complete();
      await tester.pumpAndSettle();
    });

    testWidgets('success forgets the session and says goodbye', (tester) async {
      await openDialog(tester);

      await tester.enterText(inputLabeled(_password), syntheticPassword());
      await confirm(tester);

      expect(harness.profile.calls, ['getMe', 'getMe', 'deleteAccount']);
      expect(find.byType(GoodbyeScreen), findsOneWidget);
      expect(find.text('Cuenta eliminada'), findsOneWidget);
      expect(harness.session.status, SessionStatus.signedOut);
      expect(harness.tokens.token, isNull);

      await tester.tap(find.text('Ir a iniciar sesión'));
      await tester.pumpAndSettle();

      expect(find.byType(LoginScreen), findsOneWidget);
    });
  });

  group('a token the device could not forget', () {
    testWidgets('the goodbye screen warns instead of failing silently', (
      tester,
    ) async {
      await openDialog(tester);
      harness.tokens
        ..failingDeletes = 1 << 20
        ..wipeError = PlatformException(code: 'wipe');

      await tester.enterText(inputLabeled(_password), syntheticPassword());
      await confirm(tester);
      // Past every retry of the delete and of the wipe.
      await tester.pump(const Duration(seconds: 5));
      await tester.pumpAndSettle();

      expect(
        find.textContaining('No pudimos borrar la sesión guardada'),
        findsOneWidget,
      );
    });
  });
}
