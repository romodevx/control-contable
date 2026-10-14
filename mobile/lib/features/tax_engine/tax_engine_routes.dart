import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/features/tax_engine/presentation/breakdown_screen.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_screen.dart';
import 'package:control_contable/features/tax_engine/presentation/tax_inputs_screen.dart';
import 'package:go_router/go_router.dart';

const String _taxpayerId = 'taxpayerId';
const String _period = 'period';
const String _regimeCode = 'regimeCode';
const String _fiscalYear = 'ejercicio';
const String _field = 'campo';
const String _estimatePeriod = 'periodo';

/// Locations owned by `module_tax_engine`. Ids, periods and regime codes
/// travel in the location; never an amount or the RFC.
abstract final class TaxEngineRoutes {
  /// The active taxpayer's monthly estimate (M5-25).
  static const estimate = '/estimado';

  /// The estimate opened on [period] (`YYYY-MM`; History, M8-05).
  static String estimateOf(String period) => Uri(
    path: estimate,
    queryParameters: {_estimatePeriod: period},
  ).toString();

  /// One regime's breakdown for [period] (M5-26).
  static String breakdown(String taxpayerId, String period, String regime) =>
      '$estimate/$taxpayerId/$period/regimenes/$regime';

  /// One regime's tax inputs for [fiscalYear], opened at [field] (the
  /// first missing key; M5-27).
  static String taxInputs(
    String taxpayerId,
    String regime, {
    required int fiscalYear,
    required String field,
  }) => Uri(
    path: '$estimate/$taxpayerId/datos/$regime',
    queryParameters: {_fiscalYear: '$fiscalYear', _field: field},
  ).toString();
}

/// The tax engine slice: unreachable while `module_tax_engine` is off (the
/// server's `404` is the security boundary).
final FeatureModule taxEngineFeature = FeatureModule(
  module: AppModule.taxEngine,
  routes: [
    GoRoute(
      path: TaxEngineRoutes.estimate,
      builder: (context, state) => EstimateScreen(
        initialPeriod: _periodIn(state.uri.queryParameters[_estimatePeriod]),
      ),
    ),
    GoRoute(
      path: TaxEngineRoutes.breakdown(
        ':$_taxpayerId',
        ':$_period',
        ':$_regimeCode',
      ),
      builder: (context, state) => BreakdownScreen(
        taxpayerId: state.pathParameters[_taxpayerId]!,
        period: state.pathParameters[_period]!,
        regime: state.pathParameters[_regimeCode]!,
      ),
    ),
    GoRoute(
      path: '${TaxEngineRoutes.estimate}/:$_taxpayerId/datos/:$_regimeCode',
      builder: (context, state) => TaxInputsScreen(
        taxpayerId: state.pathParameters[_taxpayerId]!,
        regime: state.pathParameters[_regimeCode]!,
        fiscalYear: int.tryParse(state.uri.queryParameters[_fiscalYear] ?? ''),
        focusKey: state.uri.queryParameters[_field],
      ),
    ),
  ],
);

/// [value] when it is a `YYYY-MM` month, else `null` (the current month).
String? _periodIn(String? value) =>
    value != null && isMonthPeriod(value) ? value : null;
