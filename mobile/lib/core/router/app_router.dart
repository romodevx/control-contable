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
/// Module slices add their `<feature>Routes` here; once the FlagsProvider
/// lands (P0-24) a module whose flag is off contributes no routes at all,
/// so its screens are unreachable and its menu entry is not rendered.
final appRouterProvider = Provider<GoRouter>((ref) {
  final router = GoRouter(
    initialLocation: AppRoutes.home,
    routes: [...homeRoutes],
    errorBuilder: (context, state) => const NotFoundScreen(),
  );
  ref.onDispose(router.dispose);
  return router;
}, name: 'appRouterProvider');
