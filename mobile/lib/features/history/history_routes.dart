import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/features/history/presentation/history_screen.dart';
import 'package:go_router/go_router.dart';

/// Locations owned by `module_history`. History is of the active taxpayer,
/// so no id travels in the location.
abstract final class HistoryRoutes {
  /// The fiscal-year list (M8-05).
  static const list = '/historial';
}

/// The History slice: unreachable while `module_history` is off (the
/// server's `404` is the security boundary).
final FeatureModule historyFeature = FeatureModule(
  module: AppModule.history,
  routes: [
    GoRoute(
      path: HistoryRoutes.list,
      builder: (context, state) => const HistoryScreen(),
    ),
  ],
);
