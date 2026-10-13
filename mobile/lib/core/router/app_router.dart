import 'package:control_contable/core/flags/flags_provider.dart';
import 'package:control_contable/core/router/boot_splash_screen.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/core/router/not_found_screen.dart';
import 'package:control_contable/core/router/router_flags.dart';
import 'package:control_contable/features/home/home_routes.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// How long a location inside a module waits on the splash for the first
/// flags answer before it fails closed to not found.
const flagsBootTimeout = Duration(seconds: 8);

/// Top-level locations. Each feature owns its own sub-paths.
abstract final class AppRoutes {
  /// The home (dashboard) location and the app's initial location.
  static const home = '/';

  /// Where a location inside a disabled module is sent.
  static const notFound = '/no-encontrado';

  /// Where a location inside a module waits while the flags load.
  static const splash = '/cargando';

  /// Query parameter of [splash] that keeps the requested location.
  static const splashFrom = 'desde';
}

/// The app's router, composed from each feature's route list.
///
/// Built once per set of feature slices, never per flag answer: a flag
/// change re-runs the redirect (`refreshListenable`) instead, so the user's
/// navigation survives a refresh that flips some other module's flag.
///
/// A location owned by a module:
/// - waits on [AppRoutes.splash] while the first flags answer is pending,
///   keeping the requested location, and returns to it once flags resolve
///   (a deep link, a notification or a post-login redirect);
/// - fails closed after [flagsBootTimeout] if the answer never arrives;
/// - redirects to [AppRoutes.notFound] while its module is disabled.
///
/// Its entry points are gated by `FlagGate`.
final appRouterProvider = Provider<GoRouter>((ref) {
  final features = ref.watch(featureModulesProvider);
  final owners = FeatureRouteOwners(features);
  final flags = RouterFlags(
    timeout: flagsBootTimeout,
    initial: ref.read(flagsProvider).value,
  );
  ref
    ..listen(flagsProvider, (_, next) {
      if (next.hasValue) flags.resolve(next.requireValue);
    })
    ..onDispose(flags.dispose);

  final router = GoRouter(
    initialLocation: AppRoutes.home,
    refreshListenable: flags,
    routes: [
      ...homeRoutes,
      GoRoute(
        path: AppRoutes.notFound,
        builder: (context, state) => const NotFoundScreen(),
      ),
      GoRoute(
        path: AppRoutes.splash,
        builder: (context, state) => const BootSplashScreen(),
      ),
      for (final feature in features) ...feature.routes,
    ],
    redirect: (context, state) => _redirect(state, owners, flags),
    errorBuilder: (context, state) => const NotFoundScreen(),
  );
  ref.onDispose(router.dispose);
  return router;
}, name: 'appRouterProvider');

String? _redirect(
  GoRouterState state,
  FeatureRouteOwners owners,
  RouterFlags flags,
) {
  if (state.matchedLocation == AppRoutes.splash) {
    return flags.isResolved ? _waitingLocation(state.uri) : null;
  }
  final owner = owners.ownerOf(state.topRoute);
  if (owner == null) return null;
  if (!flags.isResolved) {
    flags.startDeadline();
    return Uri(
      path: AppRoutes.splash,
      queryParameters: {AppRoutes.splashFrom: state.uri.toString()},
    ).toString();
  }
  return flags.isEnabled(owner) ? null : AppRoutes.notFound;
}

/// The in-app location the splash kept; home when it is missing or is not
/// an app path (never an external or protocol-relative URL).
String _waitingLocation(Uri splash) {
  final from = splash.queryParameters[AppRoutes.splashFrom];
  final isAppPath =
      from != null && from.startsWith('/') && !from.startsWith('//');
  return isAppPath ? from : AppRoutes.home;
}
