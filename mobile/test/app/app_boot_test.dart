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
}
