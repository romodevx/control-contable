import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/features/taxpayers/presentation/onboarding_screen.dart';
import 'package:control_contable/features/taxpayers/presentation/settings/add_regime_screen.dart';
import 'package:control_contable/features/taxpayers/presentation/settings/regime_options_screen.dart';
import 'package:control_contable/features/taxpayers/presentation/settings/taxpayer_settings_screen.dart';
import 'package:go_router/go_router.dart';

/// Path parameter of a taxpayer's id.
const String _taxpayerId = 'taxpayerId';

/// Path parameter of a regime's code.
const String _regimeCode = 'regimeCode';

/// Locations owned by `module_taxpayers`.
abstract final class TaxpayersRoutes {
  /// The onboarding of a new taxpayer (M2-08, M2-09). Its draft lives in
  /// memory on the screen, never in the location.
  static const onboarding = '/contribuyentes/nuevo';

  /// A taxpayer's settings (M2-10). Only the id travels in the location,
  /// never the RFC.
  static String settings(String taxpayerId) =>
      '/contribuyentes/$taxpayerId/ajustes';

  /// Adds a regime to a taxpayer.
  static String addRegime(String taxpayerId) =>
      '${settings(taxpayerId)}/regimenes/nuevo';

  /// Edits one regime's options.
  static String regimeOptions(String taxpayerId, String regimeCode) =>
      '${settings(taxpayerId)}/regimenes/$regimeCode';
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
    GoRoute(
      path: TaxpayersRoutes.settings(':$_taxpayerId'),
      builder: (context, state) => TaxpayerSettingsScreen(
        taxpayerId: state.pathParameters[_taxpayerId]!,
      ),
      routes: [
        GoRoute(
          path: 'regimenes/nuevo',
          builder: (context, state) =>
              AddRegimeScreen(taxpayerId: state.pathParameters[_taxpayerId]!),
        ),
        GoRoute(
          path: 'regimenes/:$_regimeCode',
          builder: (context, state) => RegimeOptionsScreen(
            taxpayerId: state.pathParameters[_taxpayerId]!,
            regimeCode: state.pathParameters[_regimeCode]!,
          ),
        ),
      ],
    ),
  ],
);
