import 'package:control_contable/core/session/forget_toasts_on_sign_out.dart';
import 'package:control_contable/features/home/presentation/home_screen.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../helpers/pump_app.dart';

void main() {
  testWidgets('the app boots into the placeholder home', (tester) async {
    await pumpControlContableApp(tester);

    expect(find.byType(MaterialApp), findsOneWidget);
    expect(find.byType(HomeScreen), findsOneWidget);
    expect(
      find.text('Inicio'),
      findsOneWidget,
      reason: 'the home placeholder shows its localized title',
    );
  });

  testWidgets('the root messenger forgets its toasts on sign-out', (
    tester,
  ) async {
    await pumpControlContableApp(tester);

    final guard = find.byType(ForgetToastsOnSignOut);
    expect(guard, findsOneWidget);
    expect(
      find.ancestor(of: guard, matching: find.byType(ScaffoldMessenger)),
      findsOneWidget,
      reason: 'it must sit below the root messenger it clears',
    );
  });
}
