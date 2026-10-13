import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/features/auth/presentation/register_screen.dart';
import 'package:control_contable/features/auth/presentation/verify_email_screen.dart';
import 'package:go_router/go_router.dart';

/// Locations owned by `module_auth`.
abstract final class AuthRoutes {
  /// The sign-up form.
  static const register = '/cuenta/registro';

  /// The 6-digit code screen. The email travels in memory
  /// (`pendingSignUpProvider`), never in the location.
  static const verifyEmail = '/cuenta/verificar';
}

/// The auth slice: unreachable while `module_auth` is off.
final FeatureModule authFeature = FeatureModule(
  module: AppModule.auth,
  routes: [
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
