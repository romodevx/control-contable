import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/features/tax_engine/tax_engine_routes.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';

import 'estimate_fixtures.dart';
import 'pump_estimate.dart';

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
  TaxEngineRoutes.estimate,
  TaxEngineRoutes.breakdown(fixtureTaxpayerId, '2026-03', '612'),
  TaxEngineRoutes.taxInputs(
    fixtureTaxpayerId,
    '612',
    fiscalYear: 2026,
    field: 'prior_provisional_payments',
  ),
];

/// `module_tax_engine` OFF hides every estimate surface (the server's 404
/// is the security boundary; this is UX).
void main() {
  late EstimateHarness harness;

  setUp(() => harness = EstimateHarness());

  testWidgets('the slice owns exactly the estimate, breakdown and inputs '
      'routes', (tester) async {
    expect(taxEngineFeature.module, AppModule.taxEngine);
    expect(taxEngineFeature.guestRoutes, isEmpty);
    expect(
      _paths(taxEngineFeature.allRoutes),
      unorderedEquals([
        TaxEngineRoutes.estimate,
        TaxEngineRoutes.breakdown(':taxpayerId', ':period', ':regimeCode'),
        '/estimado/:taxpayerId/datos/:regimeCode',
      ]),
    );
  });

  testWidgets('off: no entry on home, every route is not found and nothing '
      'is asked', (tester) async {
    await harness.pump(tester, enabled: false, location: null);

    expect(find.text('Ver estimado del mes'), findsNothing);
    for (final location in _locations) {
      harness.container.read(appRouterProvider).go(location);
      await tester.pumpAndSettle();

      expect(harness.location, AppRoutes.notFound, reason: location);
    }
    expect(harness.repository.calls, isEmpty);
  });

  testWidgets('on: home opens the estimate', (tester) async {
    await harness.pump(tester, location: null);

    await tester.tap(find.text('Ver estimado del mes'));
    await tester.pumpAndSettle();

    expect(harness.location, TaxEngineRoutes.estimate);
    expect(find.text('Total a pagar'), findsOneWidget);
  });

  testWidgets('on: the breakdown and the inputs form open', (tester) async {
    await harness.pump(tester, location: null);
    final titles = ['Desglose', 'Datos fiscales'];
    for (final (index, location) in _locations.skip(1).indexed) {
      harness.container.read(appRouterProvider).go(location);
      await tester.pumpAndSettle();

      expect(harness.location, location);
      expect(find.text('Página no encontrada'), findsNothing);
      expect(find.text(titles[index]), findsOneWidget);
      expect(find.text('Disponible pronto.'), findsNothing);
    }
    expect(harness.repository.breakdownCalls, hasLength(1));
    expect(harness.repository.inputsCalls, hasLength(1));
  });
}
