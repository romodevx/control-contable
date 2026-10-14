import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/features/incomes/incomes_routes.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';

import 'pump_incomes.dart';

const String _id = '0199d4a3-0000-7000-8000-000000000001';

/// Every route path of [routes], nested ones joined to their parent.
List<String> _paths(List<RouteBase> routes, [String parent = '']) => [
  for (final route in routes.whereType<GoRoute>()) ...[
    _join(parent, route.path),
    ..._paths(route.routes, _join(parent, route.path)),
  ],
];

String _join(String parent, String path) =>
    parent.isEmpty ? path : '$parent/$path';

/// `module_incomes` OFF hides every incomes screen (the server's 404 is the
/// security boundary; this is UX).
void main() {
  late IncomesHarness harness;

  setUp(() => harness = IncomesHarness());

  final locations = [
    IncomesRoutes.list,
    IncomesRoutes.create,
    IncomesRoutes.editOf(_id),
  ];

  // M3-10: the router gates the whole slice by `module_incomes`, so every
  // incomes location, nested ones included, is not found while the flag is
  // off, and nothing is loaded from the server.
  testWidgets('off: every incomes route is not found', (tester) async {
    expect(incomesFeature.module, AppModule.incomes);
    expect(incomesFeature.guestRoutes, isEmpty);
    expect(
      _paths(incomesFeature.allRoutes),
      unorderedEquals([
        IncomesRoutes.list,
        IncomesRoutes.create,
        '${IncomesRoutes.list}/:incomeId/editar',
      ]),
    );

    await harness.pump(tester, enabled: false, location: null);
    for (final location in locations) {
      harness.container.read(appRouterProvider).go(location);
      await tester.pumpAndSettle();

      expect(harness.location, AppRoutes.notFound, reason: location);
      expect(find.text('Página no encontrada'), findsOneWidget);
    }
    expect(harness.incomes.calls, isEmpty);
  });

  testWidgets('on: every incomes route opens its screen', (tester) async {
    await harness.pump(tester, location: null);
    for (final location in locations) {
      harness.container.read(appRouterProvider).go(location);
      await tester.pumpAndSettle();

      expect(harness.location, location);
      expect(find.text('Página no encontrada'), findsNothing);
    }
  });
}
