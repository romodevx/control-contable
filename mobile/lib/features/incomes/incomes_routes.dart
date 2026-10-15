import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/features/incomes/presentation/income_edit_screen.dart';
import 'package:control_contable/features/incomes/presentation/income_form_screen.dart';
import 'package:control_contable/features/incomes/presentation/incomes_screen.dart';
import 'package:go_router/go_router.dart';

/// Locations owned by `module_incomes`. They scope to the active taxpayer,
/// so no taxpayer id travels in the location (an edit carries the income's).
abstract final class IncomesRoutes {
  /// The month list (M3-08).
  static const list = '/ingresos';

  /// The month list opened on [period] (`YYYY-MM`; History, US-H6).
  static String listOf(String period) =>
      Uri(path: list, queryParameters: {recordMonthQuery: period}).toString();

  /// The new-income page (M3-09).
  static const create = '/ingresos/nuevo';

  /// The edit page of income [incomeId] (M3-09, US-I3).
  static String editOf(String incomeId) =>
      '$list/${Uri.encodeComponent(incomeId)}/editar';
}

/// The incomes slice: unreachable while `module_incomes` is off (the
/// server's `404` is the security boundary).
final FeatureModule incomesFeature = FeatureModule(
  module: AppModule.incomes,
  routes: [
    GoRoute(
      path: IncomesRoutes.list,
      builder: (context, state) => IncomesScreen(
        initialPeriod: state.uri.queryParameters[recordMonthQuery],
      ),
      routes: [
        GoRoute(
          path: 'nuevo',
          builder: (context, state) => const IncomeFormScreen(),
        ),
        GoRoute(
          path: ':incomeId/editar',
          builder: (context, state) => IncomeEditScreen(
            incomeId: state.pathParameters['incomeId']!,
            initial: state.extra is Income ? state.extra! as Income : null,
          ),
        ),
      ],
    ),
  ],
);
