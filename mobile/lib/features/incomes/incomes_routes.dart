import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/features/incomes/presentation/income_form_screen.dart';
import 'package:control_contable/features/incomes/presentation/incomes_screen.dart';
import 'package:go_router/go_router.dart';

/// Locations owned by `module_incomes`. They scope to the active taxpayer,
/// so no id travels in the location.
abstract final class IncomesRoutes {
  /// The month list (M3-08).
  static const list = '/ingresos';

  /// The new-income page (M3-09).
  static const create = '/ingresos/nuevo';
}

/// The incomes slice: unreachable while `module_incomes` is off (the
/// server's `404` is the security boundary).
final FeatureModule incomesFeature = FeatureModule(
  module: AppModule.incomes,
  routes: [
    GoRoute(
      path: IncomesRoutes.list,
      builder: (context, state) => const IncomesScreen(),
      routes: [
        GoRoute(
          path: 'nuevo',
          builder: (context, state) => const IncomeFormScreen(),
        ),
      ],
    ),
  ],
);
