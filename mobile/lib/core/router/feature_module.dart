import 'package:control_contable/core/flags/app_module.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// A module slice's contribution to the router: its routes, reachable only
/// while [module] is enabled.
///
/// The slice that signs people in (auth) also declares [signInLocation] and
/// its [guestRoutes]. While its module is on, every other location needs a
/// session (signed out: sent to [signInLocation]), and a guest route is
/// left for where the user was going once signed in.
@immutable
class FeatureModule {
  /// Declares [routes] and [guestRoutes] owned by [module].
  const new({
    required this.module,
    this.routes = const [],
    this.guestRoutes = const [],
    this.signInLocation,
  });

  /// The flag that decides whether the module's routes can be reached.
  final AppModule module;

  /// The module's top-level routes; they need a session when some slice
  /// signs people in.
  final List<RouteBase> routes;

  /// Top-level routes for people not signed in (login, sign-up).
  final List<RouteBase> guestRoutes;

  /// Where a signed-out user is sent; `null` for a slice that does not
  /// sign people in.
  final String? signInLocation;

  /// Every top-level route of the slice.
  List<RouteBase> get allRoutes => [...routes, ...guestRoutes];
}

/// Which module owns each route of the given feature slices, at any depth
/// (nested and shell routes included).
class FeatureRouteOwners {
  /// Indexes every route under each of [features].
  new(Iterable<FeatureModule> features) {
    for (final feature in features) {
      _index(feature.allRoutes, feature.module);
      _indexGuests(feature.guestRoutes);
    }
  }

  // Identity: two routes with the same path are still different routes.
  final Map<RouteBase, AppModule> _owners = Map.identity();
  final Set<RouteBase> _guests = Set.identity();

  void _indexGuests(List<RouteBase> routes) {
    for (final route in routes) {
      _guests.add(route);
      _indexGuests(route.routes);
    }
  }

  void _index(List<RouteBase> routes, AppModule module) {
    for (final route in routes) {
      _owners[route] = module;
      _index(route.routes, module);
    }
  }

  /// The module that owns [route], or `null` for a route outside every
  /// feature slice (home, not found).
  AppModule? ownerOf(RouteBase? route) => route == null ? null : _owners[route];

  /// Whether [route] is for people not signed in (a guest route, at any
  /// depth).
  bool isGuest(RouteBase? route) => route != null && _guests.contains(route);
}

/// Every module slice in this build. A slice adds its `FeatureModule` here
/// (M1 onwards); `appRouterProvider` lets the user reach only enabled ones.
final featureModulesProvider = Provider<List<FeatureModule>>(
  (ref) => const [],
  name: 'featureModulesProvider',
);
