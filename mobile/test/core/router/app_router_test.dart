import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/core/router/not_found_screen.dart';
import 'package:control_contable/features/home/presentation/home_screen.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';

import '../../helpers/pump_app.dart';

GoRouter _routerOf(WidgetTester tester) {
  final container = ProviderScope.containerOf(
    tester.element(find.byType(MaterialApp)),
  );
  return container.read(appRouterProvider);
}

void main() {
  testWidgets('the initial location is the home route', (tester) async {
    await pumpControlContableApp(tester);

    final router = _routerOf(tester);
    expect(router.routerDelegate.currentConfiguration.uri.path, AppRoutes.home);
    expect(find.byType(HomeScreen), findsOneWidget);
  });

  testWidgets('an unknown location shows the not-found screen', (tester) async {
    await pumpControlContableApp(tester);

    _routerOf(tester).go('/no-existe');
    await tester.pumpAndSettle();

    expect(find.byType(NotFoundScreen), findsOneWidget);
    expect(find.text('Página no encontrada'), findsOneWidget);

    await tester.tap(find.text('Ir al inicio'));
    await tester.pumpAndSettle();
    expect(find.byType(HomeScreen), findsOneWidget);
  });
}
