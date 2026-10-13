import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flags_provider.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/core/router/not_found_screen.dart';
import 'package:control_contable/features/home/home_routes.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// Top-level locations. Each feature owns its own sub-paths.
abstract final class AppRoutes {
  /// The home (dashboard) location and the app's initial location.
  static const home = '/';

  /// Where a location inside a disabled module is sent.
  static const notFound = '/no-encontrado';
}

/// The app's router, composed from each feature's route list.
///
/// Built once per set of feature slices, never per flag answer: a flag
/// change re-runs the redirect (`refreshListenable`) instead, so the user's
/// navigation survives a refresh that flips some other module's flag. A
/// location owned by a disabled module redirects to [AppRoutes.notFound];
/// its entry points are gated by `FlagGate`.
final appRouterProvider = Provider<GoRouter>((ref) {
  final features = ref.watch(featureModulesProvider);
  final owners = FeatureRouteOwners(features);
  final enabled = ValueNotifier<Set<AppModule>>(
    ref.read(flagsProvider).value ?? const {},
  );
  ref
    ..listen(flagsProvider, (_, next) => enabled.value = next.value ?? const {})
    ..onDispose(enabled.dispose);

  final router = GoRouter(
    initialLocation: AppRoutes.home,
    refreshListenable: enabled,
    routes: [
      ...homeRoutes,
      GoRoute(
        path: AppRoutes.notFound,
        builder: (context, state) => const NotFoundScreen(),
      ),
      for (final feature in features) ...feature.routes,
    ],
    redirect: (context, state) {
      final owner = owners.ownerOf(state.topRoute);
      if (owner == null || enabled.value.contains(owner)) return null;
      return AppRoutes.notFound;
    },
    errorBuilder: (context, state) => const NotFoundScreen(),
  );
  ref.onDispose(router.dispose);
  return router;
}, name: 'appRouterProvider');
