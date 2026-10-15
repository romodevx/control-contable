import 'package:control_contable/api/generated/models/annual_computed_section_kind.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/core/router/not_found_screen.dart';
import 'package:control_contable/features/annual/presentation/annual_breakdown_screen.dart';
import 'package:control_contable/features/annual/presentation/annual_screen.dart';
import 'package:go_router/go_router.dart';

const String _taxpayerId = 'taxpayerId';
const String _fiscalYear = 'ejercicio';
const String _section = 'seccion';
const String _root = '/anual';
const String _year = '$_root/:$_taxpayerId/:$_fiscalYear';

/// Locations owned by `module_annual`. Ids, years and section kinds travel
/// in the location; never an amount or the RFC. The input screens arrive
/// with ANN-19..21, each with its own location: none is a placeholder.
abstract final class AnnualRoutes {
  /// The active taxpayer's annual estimate (ANN-17).
  static const String estimate = _root;

  /// One section's computation trace (ANN-18).
  static String breakdown(
    String taxpayerId,
    int fiscalYear,
    AnnualComputedSectionKind section,
  ) => '${_of(taxpayerId, fiscalYear)}/secciones/${section.json}';

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
  ],
);

int? _yearOf(GoRouterState state) =>
    int.tryParse(state.pathParameters[_fiscalYear] ?? '');
