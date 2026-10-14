import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/features/auth/presentation/login_screen.dart';
import 'package:control_contable/features/auth/presentation/register_screen.dart';
import 'package:control_contable/features/auth/presentation/verify_email_screen.dart';
import 'package:go_router/go_router.dart';

/// Locations owned by `module_auth`.
abstract final class AuthRoutes {
  /// The login form: where a signed-out user is sent.
  static const login = '/cuenta/entrar';

  /// The sign-up form.
  static const register = '/cuenta/registro';

  /// The 6-digit code screen. The email travels in memory
  /// (`pendingSignUpProvider`), never in the location.
  static const verifyEmail = '/cuenta/verificar';
}

/// The auth slice: unreachable while `module_auth` is off. While it is on,
/// the rest of the app needs a session and these screens are for guests.
final FeatureModule authFeature = FeatureModule(
  module: AppModule.auth,
  signInLocation: AuthRoutes.login,
  guestRoutes: [
    GoRoute(
      path: AuthRoutes.login,
      builder: (context, state) => const LoginScreen(),
    ),
    GoRoute(
      path: AuthRoutes.register,
      builder: (context, state) => const RegisterScreen(),
    ),
    GoRoute(
      path: AuthRoutes.verifyEmail,
      builder: (context, state) => const VerifyEmailScreen(),
    ),
  ],
);
