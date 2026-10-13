import 'package:control_contable/core/flags/app_module.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// A module slice's contribution to the router: its routes, registered only
/// while [module] is enabled.
@immutable
class FeatureModule {
  /// Declares [routes] owned by [module].
  const new({required this.module, required this.routes});

  /// The flag that decides whether [routes] exist.
  final AppModule module;

  /// The module's top-level routes.
  final List<RouteBase> routes;
}

/// Every module slice in this build. A slice adds its `FeatureModule` here
/// (M1 onwards); `appRouterProvider` keeps only the enabled ones.
final featureModulesProvider = Provider<List<FeatureModule>>(
  (ref) => const [],
  name: 'featureModulesProvider',
);
