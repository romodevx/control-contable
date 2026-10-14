import 'package:control_contable/core/flags/app_module.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'pump_dashboard.dart';

const String _placeholder = 'Aquí aparecerá el estimado del mes.';

/// The dashboard's own widgets: none may show while `module_dashboard` is
/// OFF.
const List<String> _dashboardKeys = [
  'dashboard-hero',
  'dashboard-month',
  'dashboard-status',
];

/// `module_dashboard` replaces Home (M6-05). OFF, the app keeps the previous
/// Home exactly, with its own flag-gated entries; the server's 404 is the
/// security boundary, this is UX.
void main() {
  late DashboardHarness harness;

  setUp(() => harness = DashboardHarness());

  testWidgets('OFF: the previous Home with every entry of the enabled '
      'modules, no dashboard widget and no dashboard read', (tester) async {
    await harness.pump(
      tester,
      modules: homeModules.difference({AppModule.dashboard}),
    );

    expect(harness.location, '/');
    expect(find.text(_placeholder), findsOneWidget);
    expect(find.text('Ver estimado del mes'), findsOneWidget);
    expect(find.text('Ingresos'), findsOneWidget);
    expect(find.text('Gastos'), findsOneWidget);
    for (final key in _dashboardKeys) {
      expect(find.byKey(ValueKey(key)), findsNothing, reason: key);
    }
    expect(harness.repository.calls, isEmpty);
  });

  testWidgets('OFF with the engine OFF too: the previous Home says nothing '
      'about an estimate', (tester) async {
    await harness.pump(
      tester,
      modules: homeModules.difference({
        AppModule.dashboard,
        AppModule.taxEngine,
      }),
    );

    expect(
      find.textContaining(RegExp('estimad', caseSensitive: false)),
      findsNothing,
    );
    expect(find.text('Ingresos'), findsOneWidget);
    expect(harness.repository.calls, isEmpty);
  });

  testWidgets('ON: the dashboard replaces the previous Home', (tester) async {
    tester.view
      ..physicalSize = const Size(400, 3000)
      ..devicePixelRatio = 1;
    addTearDown(tester.view.reset);
    await harness.pump(tester);

    expect(find.text(_placeholder), findsNothing);
    expect(find.text('Ver estimado del mes'), findsNothing);
    expect(find.byKey(const ValueKey('dashboard-hero')), findsOneWidget);
    expect(harness.repository.calls, hasLength(1));
  });
}
