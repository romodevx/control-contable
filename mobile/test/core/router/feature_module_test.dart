import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';

Widget _blank(BuildContext context, GoRouterState state) =>
    const SizedBox.shrink();

void main() {
  final detail = GoRoute(path: 'detalle', builder: _blank);
  final inShell = GoRoute(path: '/en-shell', builder: _blank);
  final incomes = GoRoute(path: '/ingresos', builder: _blank, routes: [detail]);
  final owners = FeatureRouteOwners([
    FeatureModule(module: AppModule.incomes, routes: [incomes]),
    FeatureModule(
      module: AppModule.expenses,
      routes: [
        ShellRoute(
          builder: (context, state, child) => child,
          routes: [inShell],
        ),
      ],
    ),
  ]);

  test('a nested route belongs to the module of its top-level route', () {
    expect(owners.ownerOf(incomes), AppModule.incomes);
    expect(owners.ownerOf(detail), AppModule.incomes);
    expect(owners.ownerOf(inShell), AppModule.expenses);
  });

  test('a route outside every feature slice has no owner', () {
    expect(owners.ownerOf(GoRoute(path: '/', builder: _blank)), isNull);
    expect(owners.ownerOf(null), isNull);
  });
}
