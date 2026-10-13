import 'package:control_contable/features/auth/auth_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'pump_auth.dart';

/// `module_auth` OFF hides every auth screen (the server's 404 is the
/// security boundary; this is UX).
void main() {
  late AuthHarness harness;

  setUp(() => harness = AuthHarness());

  Finder createAccount() => find.widgetWithText(OutlinedButton, 'Crear cuenta');

  for (final location in [AuthRoutes.register, AuthRoutes.verifyEmail]) {
    testWidgets('off: $location is not found', (tester) async {
      await harness.pump(tester, location: location, enabled: false);

      expect(find.text('Página no encontrada'), findsOneWidget);
      expect(harness.repository.documentCalls, 0);
    });
  }

  testWidgets('off: home offers no way to create an account', (tester) async {
    await harness.pump(tester, enabled: false);

    expect(createAccount(), findsNothing);
  });

  testWidgets('on: home offers "Crear cuenta", which opens the form', (
    tester,
  ) async {
    await harness.pump(tester);

    await tester.tap(createAccount());
    await tester.pumpAndSettle();

    expect(find.text('Correo electrónico'), findsOneWidget);
  });
}
