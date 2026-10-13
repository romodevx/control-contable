import 'package:control_contable/core/flags/app_module.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// Placeholder: implemented after the failing test (TDD).
class FeatureModule {
  const new({required this.module, required this.routes});

  final AppModule module;
  final List<RouteBase> routes;
}

final featureModulesProvider = Provider<List<FeatureModule>>(
  (ref) => const [],
);
