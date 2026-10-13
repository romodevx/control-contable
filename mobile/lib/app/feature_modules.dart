import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/features/auth/auth_routes.dart';

/// Every module slice this build ships, composed at the app root so `core/`
/// never imports a feature. The router lets the user reach only the enabled
/// ones (`featureModulesProvider`).
final List<FeatureModule> appFeatureModules = [authFeature];
