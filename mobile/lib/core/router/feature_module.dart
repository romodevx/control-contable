import 'package:control_contable/core/flags/app_module.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// A module slice's contribution to the router: its routes, reachable only
/// while [module] is enabled.
@immutable
class FeatureModule {
  /// Declares [routes] owned by [module].
  const new({required this.module, required this.routes});

  /// The flag that decides whether [routes] can be reached.
  final AppModule module;

  /// The module's top-level routes.
  final List<RouteBase> routes;
}

/// Which module owns each route of the given feature slices, at any depth
/// (nested and shell routes included).
class FeatureRouteOwners {
  /// Indexes every route under each of [features].
  new(Iterable<FeatureModule> features) {
    for (final feature in features) {
      _index(feature.routes, feature.module);
    }
  }

  // Identity: two routes with the same path are still different routes.
  final Map<RouteBase, AppModule> _owners = Map.identity();

  void _index(List<RouteBase> routes, AppModule module) {
    for (final route in routes) {
      _owners[route] = module;
      _index(route.routes, module);
    }
  }

  /// The module that owns [route], or `null` for a route outside every
  /// feature slice (home, not found).
  AppModule? ownerOf(RouteBase? route) => route == null ? null : _owners[route];
}

/// Every module slice in this build. A slice adds its `FeatureModule` here
/// (M1 onwards); `appRouterProvider` lets the user reach only enabled ones.
final featureModulesProvider = Provider<List<FeatureModule>>(
  (ref) => const [],
  name: 'featureModulesProvider',
);
