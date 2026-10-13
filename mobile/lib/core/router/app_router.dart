import 'package:control_contable/core/flags/flags_provider.dart';
import 'package:control_contable/core/router/boot_splash_screen.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/core/router/not_found_screen.dart';
import 'package:control_contable/core/router/router_flags.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/home/home_routes.dart';
import 'package:flutter/foundation.dart';
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

  /// Query parameter of [splash] (and of the sign-in location) that keeps
  /// the requested location.
  static const splashFrom = 'desde';
}

/// The app's router, composed from each feature's route list.
///
/// Built once per set of feature slices, never per flag answer: a flag or
/// session change re-runs the redirect (`refreshListenable`) instead, so
/// the user's navigation survives a refresh that flips some other module's
/// flag.
///
/// A location owned by a module:
/// - waits on [AppRoutes.splash] while the first flags answer is pending,
///   keeping the requested location, and returns to it once flags resolve
///   (a deep link, a notification or a post-login redirect);
/// - fails closed after [flagsBootTimeout] if the answer never arrives;
/// - redirects to [AppRoutes.notFound] while its module is disabled.
///
/// While the slice that signs people in is enabled, the session guards
/// apply too (see [_Guards]). Entry points are gated by `FlagGate`.
final appRouterProvider = Provider<GoRouter>((ref) {
  final features = ref.watch(featureModulesProvider);
  final session = ref.watch(sessionProvider);
  final flags = RouterFlags(
    timeout: flagsBootTimeout,
    initial: ref.read(flagsProvider).value,
  );
  ref
    ..listen(flagsProvider, (_, next) {
      if (next.hasValue) flags.resolve(next.requireValue);
    })
    ..onDispose(flags.dispose);
  final guards = _Guards(
    owners: FeatureRouteOwners(features),
    flags: flags,
    session: session,
    signIn: features.where((f) => f.signInLocation != null).firstOrNull,
  );

  final router = GoRouter(
    initialLocation: AppRoutes.home,
    refreshListenable: Listenable.merge([flags, session]),
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
      for (final feature in features) ...feature.allRoutes,
    ],
    redirect: (context, state) => guards.redirect(state),
    errorBuilder: (context, state) => const NotFoundScreen(),
  );
  ref.onDispose(router.dispose);
  return router;
}, name: 'appRouterProvider');

/// The router's redirect rules: module flags first (a disabled module is
/// not found), then, while the [signIn] slice's module is on, the session:
/// - while the stored session is being read, the location waits on the
///   splash;
/// - signed out, any location but a guest route goes to the sign-in
///   location, keeping the requested one ([AppRoutes.splashFrom]);
/// - signed in, a guest route goes on to the kept location (or home).
///
/// The server's 401/404 is the security boundary; this is UX.
class _Guards {
  new({
    required this.owners,
    required this.flags,
    required this.session,
    required this.signIn,
  });

  final FeatureRouteOwners owners;
  final RouterFlags flags;
  final Session session;
  final FeatureModule? signIn;

  String? redirect(GoRouterState state) {
    final location = state.matchedLocation;
    if (location == AppRoutes.splash) return _leaveSplash(state.uri);
    if (location == AppRoutes.notFound) return null;
    final owner = owners.ownerOf(state.topRoute);
    final waitsForFlags = owner != null || signIn != null;
    if (waitsForFlags && !flags.isResolved) return _wait(state.uri);
    if (owner != null && !flags.isEnabled(owner)) return AppRoutes.notFound;
    return _sessionRedirect(state);
  }

  /// Leaves the splash once flags (and, when guarded, the session) are known.
  String? _leaveSplash(Uri splash) {
    if (!flags.isResolved) {
      // Reached directly (restored state, a stale link): the wait is bounded.
      flags.startDeadline();
      return null;
    }
    if (_isGuarded && session.status == SessionStatus.restoring) return null;
    return _keptLocation(splash);
  }

  String? _sessionRedirect(GoRouterState state) {
    final signInLocation = signIn?.signInLocation;
    if (!_isGuarded || signInLocation == null) return null;
    if (session.status == SessionStatus.restoring) return _wait(state.uri);
    final isGuestRoute = owners.isGuest(state.topRoute);
    if (session.isSignedIn) {
      return isGuestRoute ? _keptLocation(state.uri) : null;
    }
    return isGuestRoute ? null : _withKept(signInLocation, state.uri);
  }

  bool get _isGuarded {
    final module = signIn?.module;
    return module != null && flags.isEnabled(module);
  }

  String _wait(Uri requested) {
    flags.startDeadline();
    return _withKept(AppRoutes.splash, requested);
  }
}

/// [path], keeping [requested] to return to later (home needs no keeping).
String _withKept(String path, Uri requested) {
  final isHome = requested.path == AppRoutes.home && !requested.hasQuery;
  return Uri(
    path: path,
    queryParameters: isHome
        ? null
        : {AppRoutes.splashFrom: requested.toString()},
  ).toString();
}

/// The in-app location kept in [uri]; home when it is missing or is not an
/// app path (never an external or protocol-relative URL).
String _keptLocation(Uri uri) {
  final from = uri.queryParameters[AppRoutes.splashFrom];
  final isAppPath =
      from != null && from.startsWith('/') && !from.startsWith('//');
  return isAppPath ? from : AppRoutes.home;
}
