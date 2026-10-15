import 'package:control_contable/api/generated/models/annual_computed_section_kind.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/features/annual/annual_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';

import '../tax_engine/estimate_fixtures.dart' show fixtureTaxpayerId;
import 'pump_annual.dart';

/// Every route path of [routes], nested ones joined to their parent.
List<String> _paths(List<RouteBase> routes, [String parent = '']) => [
  for (final route in routes.whereType<GoRoute>()) ...[
    _join(parent, route.path),
    ..._paths(route.routes, _join(parent, route.path)),
  ],
];

String _join(String parent, String path) =>
    parent.isEmpty ? path : '$parent/$path';

final List<String> _locations = [
  AnnualRoutes.estimate,
  AnnualRoutes.breakdown(
    fixtureTaxpayerId,
    2026,
    AnnualComputedSectionKind.pfArt152,
  ),
  AnnualRoutes.personalDeductions(fixtureTaxpayerId, 2026),
  AnnualRoutes.salaryStatements(fixtureTaxpayerId, 2026),
  AnnualRoutes.annualInputs(fixtureTaxpayerId, 2026),
  AnnualRoutes.filedPayments(fixtureTaxpayerId, 2026, '612'),
];

/// `module_annual` OFF hides every annual surface (the server's 404 is the
/// security boundary; this is UX).
void main() {
  late AnnualHarness harness;

  setUp(() => harness = AnnualHarness());

  test('the slice owns exactly the annual routes', () {
    expect(annualFeature.module, AppModule.annual);
    expect(annualFeature.guestRoutes, isEmpty);
    expect(
      _paths(annualFeature.allRoutes),
      unorderedEquals([
        '/anual',
        '/anual/:taxpayerId/:ejercicio/secciones/:seccion',
        '/anual/:taxpayerId/:ejercicio/deducciones',
        '/anual/:taxpayerId/:ejercicio/constancias',
        '/anual/:taxpayerId/:ejercicio/datos',
        '/anual/:taxpayerId/:ejercicio/pagos/:regimeCode',
      ]),
    );
  });

  testWidgets('off: no entry on home, every route is not found and nothing '
      'is asked', (tester) async {
    await harness.pump(tester, enabled: false, location: null);

    expect(find.byTooltip('Declaración anual'), findsNothing);
    for (final location in _locations) {
      harness.container.read(appRouterProvider).go(location);
      await tester.pumpAndSettle();

      expect(harness.location, AppRoutes.notFound, reason: location);
    }
    expect(harness.repository.calls, isEmpty);
    expect(harness.repository.breakdownCalls, isEmpty);
  });

  testWidgets('on: home opens the annual estimate', (tester) async {
    await harness.pump(tester, location: null);

    await tester.tap(find.byTooltip('Declaración anual'));
    await tester.pumpAndSettle();

    expect(harness.location, AnnualRoutes.estimate);
    expect(find.text('Saldo a favor'), findsWidgets);
  });

  testWidgets('on: every annual route opens', (tester) async {
    await harness.pump(tester, location: null);
    for (final location in _locations) {
      harness.container.read(appRouterProvider).go(location);
      await tester.pumpAndSettle();

      expect(harness.location, location);
      expect(find.text('Página no encontrada'), findsNothing);
      expect(find.byType(Scaffold), findsWidgets);
    }
  });
}
