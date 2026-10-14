import 'package:control_contable/features/auth/auth_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'pump_auth.dart';

/// `module_auth` OFF hides every auth screen (the server's 404 is the
/// security boundary; this is UX).
void main() {
  late AuthHarness harness;

  setUp(() => harness = AuthHarness());

  Finder createAccount() => find.widgetWithText(TextButton, 'Crear cuenta');

  for (final location in [
    AuthRoutes.login,
    AuthRoutes.register,
    AuthRoutes.verifyEmail,
  ]) {
    testWidgets('off: $location is not found', (tester) async {
      await harness.pump(tester, location: location, enabled: false);

      expect(find.text('Página no encontrada'), findsOneWidget);
      expect(harness.repository.documentCalls, 0);
    });
  }

  testWidgets('off: the app opens on home, with no way to sign in', (
    tester,
  ) async {
    await harness.pump(tester, enabled: false);

    expect(find.text('Inicio'), findsOneWidget);
    expect(createAccount(), findsNothing);
  });

  testWidgets('on: login offers "Crear cuenta", which opens the form', (
    tester,
  ) async {
    await harness.pump(tester);

    await tester.tap(createAccount());
    await tester.pumpAndSettle();

    expect(find.text('Correo electrónico'), findsOneWidget);
  });
}
