import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/features/taxpayers/taxpayers_routes.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';

import 'pump_onboarding.dart';

const String _id = '0192f0c4-7b1e-7c3a-9d2e-00000000000d';

/// Every route path of [routes], nested ones joined to their parent.
List<String> _paths(List<RouteBase> routes, [String parent = '']) => [
  for (final route in routes.whereType<GoRoute>()) ...[
    _join(parent, route.path),
    ..._paths(route.routes, _join(parent, route.path)),
  ],
];

String _join(String parent, String path) =>
    parent.isEmpty ? path : '$parent/$path';

/// `module_taxpayers` OFF hides every taxpayers screen (the server's 404 is
/// the security boundary; this is UX).
void main() {
  late OnboardingHarness harness;

  setUp(() => harness = OnboardingHarness());

  // M2-11: the router gates the whole slice by `module_taxpayers`, so every
  // taxpayers location, nested ones included, is not found while the flag is
  // off, and nothing is loaded from the server.
  testWidgets('off: every taxpayers route is not found', (tester) async {
    expect(taxpayersFeature.module, AppModule.taxpayers);
    expect(taxpayersFeature.guestRoutes, isEmpty);
    expect(
      _paths(taxpayersFeature.allRoutes),
      unorderedEquals([
        TaxpayersRoutes.onboarding,
        TaxpayersRoutes.settings(':taxpayerId'),
        TaxpayersRoutes.addRegime(':taxpayerId'),
        TaxpayersRoutes.regimeOptions(':taxpayerId', ':regimeCode'),
      ]),
    );
    final locations = [
      TaxpayersRoutes.onboarding,
      TaxpayersRoutes.settings(_id),
      TaxpayersRoutes.addRegime(_id),
      TaxpayersRoutes.regimeOptions(_id, '612'),
    ];

    await harness.pump(tester, enabled: false, location: null);
    for (final location in locations) {
      harness.container.read(appRouterProvider).go(location);
      await tester.pumpAndSettle();

      expect(harness.location, AppRoutes.notFound, reason: location);
      expect(find.text('Página no encontrada'), findsOneWidget);
    }
    expect(harness.repository.calls, isEmpty);
    expect(harness.repository.catalogCalls, isEmpty);
    expect(harness.repository.creates, isEmpty);
  });

  testWidgets('on: every taxpayers route opens its screen', (tester) async {
    final locations = [
      TaxpayersRoutes.onboarding,
      TaxpayersRoutes.settings(_id),
      TaxpayersRoutes.addRegime(_id),
      TaxpayersRoutes.regimeOptions(_id, '612'),
    ];

    await harness.pump(tester, location: null);
    for (final location in locations) {
      harness.container.read(appRouterProvider).go(location);
      await tester.pumpAndSettle();

      expect(harness.location, location);
      expect(find.text('Página no encontrada'), findsNothing);
    }
  });
}
