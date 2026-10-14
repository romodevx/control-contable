import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/features/taxpayers/presentation/onboarding_screen.dart';
import 'package:go_router/go_router.dart';

/// Locations owned by `module_taxpayers`.
abstract final class TaxpayersRoutes {
  /// The onboarding of a new taxpayer (M2-08, M2-09). Its draft lives in
  /// memory on the screen, never in the location.
  static const onboarding = '/contribuyentes/nuevo';
}

/// The taxpayers slice: unreachable while `module_taxpayers` is off (the
/// server's `404` is the security boundary).
final FeatureModule taxpayersFeature = FeatureModule(
  module: AppModule.taxpayers,
  routes: [
    GoRoute(
      path: TaxpayersRoutes.onboarding,
      builder: (context, state) => const OnboardingScreen(),
    ),
  ],
);
