import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/features/auth/presentation/goodbye_screen.dart';
import 'package:control_contable/features/auth/presentation/login_screen.dart';
import 'package:control_contable/features/auth/presentation/profile_screen.dart';
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

  /// *Mi cuenta*: profile, legal documents and account deletion.
  static const profile = '/cuenta/perfil';

  /// Shown after the account is deleted (the user is signed out by then).
  static const goodbye = '/cuenta/adios';
}

/// The auth slice: unreachable while `module_auth` is off. While it is on,
/// the rest of the app needs a session and the guest screens are for
/// people not signed in.
final FeatureModule authFeature = FeatureModule(
  module: AppModule.auth,
  signInLocation: AuthRoutes.login,
  routes: [
    GoRoute(
      path: AuthRoutes.profile,
      builder: (context, state) => const ProfileScreen(),
    ),
  ],
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
    GoRoute(
      path: AuthRoutes.goodbye,
      builder: (context, state) => const GoodbyeScreen(),
    ),
  ],
);
