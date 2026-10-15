import 'package:control_contable/api/generated/models/annual_computed_section_kind.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/core/router/not_found_screen.dart';
import 'package:control_contable/features/annual/presentation/annual_breakdown_screen.dart';
import 'package:control_contable/features/annual/presentation/annual_input_pending_screen.dart';
import 'package:control_contable/features/annual/presentation/annual_screen.dart';
import 'package:control_contable/features/annual/presentation/personal_deduction_form_page.dart';
import 'package:control_contable/features/annual/presentation/personal_deductions_screen.dart';
import 'package:flutter/widgets.dart';
import 'package:go_router/go_router.dart';

const String _taxpayerId = 'taxpayerId';
const String _fiscalYear = 'ejercicio';
const String _section = 'seccion';
const String _regimeCode = 'regimeCode';
const String _root = '/anual';
const String _year = '$_root/:$_taxpayerId/:$_fiscalYear';

/// Locations owned by `module_annual`. Ids, years, section kinds and regime
/// codes travel in the location; never an amount or the RFC.
abstract final class AnnualRoutes {
  /// The active taxpayer's annual estimate (ANN-17).
  static const String estimate = _root;

  /// One section's computation trace (ANN-18).
  static String breakdown(
    String taxpayerId,
    int fiscalYear,
    AnnualComputedSectionKind section,
  ) => '${_of(taxpayerId, fiscalYear)}/secciones/${section.json}';

  /// The personal deductions of the year (ANN-19).
  static String personalDeductions(String taxpayerId, int fiscalYear) =>
      '${_of(taxpayerId, fiscalYear)}/deducciones';

  /// The new personal deduction form of the year (ANN-19).
  static String newPersonalDeduction(String taxpayerId, int fiscalYear) =>
      '${personalDeductions(taxpayerId, fiscalYear)}/nueva';

  /// The salary constancias of the year (ANN-20).
  static String salaryStatements(String taxpayerId, int fiscalYear) =>
      '${_of(taxpayerId, fiscalYear)}/constancias';

  /// The per-year annual facts (ANN-21).
  static String annualInputs(String taxpayerId, int fiscalYear) =>
      '${_of(taxpayerId, fiscalYear)}/datos';

  /// The provisional payments filed for [regime] in the year (ANN-21).
  static String filedPayments(
    String taxpayerId,
    int fiscalYear,
    String regime,
  ) => '${_of(taxpayerId, fiscalYear)}/pagos/$regime';

  static String _of(String taxpayerId, int fiscalYear) =>
      '$_root/$taxpayerId/$fiscalYear';
}

/// The annual slice: unreachable while `module_annual` is off (the server's
/// `404` is the security boundary; the server also turns it off with
/// `module_tax_engine`).
final FeatureModule annualFeature = FeatureModule(
  module: AppModule.annual,
  routes: [
    GoRoute(
      path: AnnualRoutes.estimate,
      builder: (context, state) => const AnnualScreen(),
    ),
    GoRoute(
      path: '$_year/secciones/:$_section',
      builder: (context, state) {
        final year = _yearOf(state);
        final section = AnnualComputedSectionKind.fromJson(
          state.pathParameters[_section]!,
        );
        if (year == null || section == AnnualComputedSectionKind.$unknown) {
          return const NotFoundScreen();
        }
        return AnnualBreakdownScreen(
          taxpayerId: state.pathParameters[_taxpayerId]!,
          fiscalYear: year,
          section: section,
        );
      },
    ),
    GoRoute(
      path: '$_year/deducciones',
      builder: (context, state) => _ofYear(
        state,
        (taxpayerId, year) =>
            PersonalDeductionsScreen(taxpayerId: taxpayerId, fiscalYear: year),
      ),
    ),
    GoRoute(
      path: '$_year/deducciones/nueva',
      builder: (context, state) => _ofYear(
        state,
        (taxpayerId, year) =>
            PersonalDeductionFormPage(taxpayerId: taxpayerId, fiscalYear: year),
      ),
    ),
    for (final (path, input) in [
      ('$_year/constancias', AnnualInput.salaryStatements),
      ('$_year/datos', AnnualInput.annualInputs),
      ('$_year/pagos/:$_regimeCode', AnnualInput.filedPayments),
    ])
      GoRoute(path: path, builder: (context, state) => _pending(state, input)),
  ],
);

/// Until ANN-19..21 ship their screens, each input location says so.
Widget _pending(GoRouterState state, AnnualInput input) =>
    _yearOf(state) == null
    ? const NotFoundScreen()
    : AnnualInputPendingScreen(input: input);

/// The screen [build] makes for the location's taxpayer and year, or not
/// found when the year is not a number.
Widget _ofYear(
  GoRouterState state,
  Widget Function(String taxpayerId, int fiscalYear) build,
) => switch (_yearOf(state)) {
  final year? => build(state.pathParameters[_taxpayerId]!, year),
  null => const NotFoundScreen(),
};

int? _yearOf(GoRouterState state) =>
    int.tryParse(state.pathParameters[_fiscalYear] ?? '');
