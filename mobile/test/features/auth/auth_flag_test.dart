import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/features/auth/auth_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';

import 'pump_auth.dart';

/// `module_auth` OFF hides every auth screen (the server's 404 is the
/// security boundary; this is UX).
void main() {
  late AuthHarness harness;

  setUp(() => harness = AuthHarness());

  Finder createAccount() => find.widgetWithText(TextButton, 'Crear cuenta');

  // M1-14: the router gates the whole slice by `module_auth`, so every auth
  // location (guest or signed-in) is not found while the flag is off.
  testWidgets('off: every auth route is not found', (tester) async {
    final paths = [
      for (final route in authFeature.allRoutes) (route as GoRoute).path,
    ];
    expect(authFeature.module, AppModule.auth);
    expect(
      paths,
      unorderedEquals([
        AuthRoutes.login,
        AuthRoutes.register,
        AuthRoutes.verifyEmail,
        AuthRoutes.profile,
        AuthRoutes.goodbye,
      ]),
    );

    await harness.pump(tester, enabled: false, signedIn: true);
    for (final path in paths) {
      harness.router.go(path);
      await tester.pumpAndSettle();

      expect(find.text('Página no encontrada'), findsOneWidget, reason: path);
    }
    expect(harness.repository.documentCalls, 0);
    expect(harness.profile.calls, isEmpty);
  });

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
