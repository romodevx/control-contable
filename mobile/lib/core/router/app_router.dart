import 'package:control_contable/core/flags/flags_provider.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/core/router/not_found_screen.dart';
import 'package:control_contable/features/home/home_routes.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// Top-level locations. Each feature owns its own sub-paths.
abstract final class AppRoutes {
  /// The home (dashboard) location and the app's initial location.
  static const home = '/';
}

/// The app's router, composed from each feature's route list.
///
/// A module whose flag is off contributes no routes at all, so its screens
/// are unreachable (not found) and its entry points are gated by `FlagGate`.
/// The router is rebuilt only when a module's flag actually changes.
final appRouterProvider = Provider<GoRouter>((ref) {
  final enabledRoutes = [
    for (final feature in ref.watch(featureModulesProvider))
      if (ref.watch(moduleEnabledProvider(feature.module))) ...feature.routes,
  ];
  final router = GoRouter(
    initialLocation: AppRoutes.home,
    routes: [...homeRoutes, ...enabledRoutes],
    errorBuilder: (context, state) => const NotFoundScreen(),
  );
  ref.onDispose(router.dispose);
  return router;
}, name: 'appRouterProvider');
