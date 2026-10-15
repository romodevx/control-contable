import 'package:control_contable/api/generated/models/annual_input_target.dart';
import 'package:control_contable/api/generated/models/annual_missing_input.dart';
import 'package:control_contable/features/annual/annual_routes.dart';
import 'package:control_contable/features/tax_engine/tax_engine_routes.dart';

/// Where [input] is completed for [taxpayerId] in [fiscalYear]: its annual
/// screen, or the monthly tax-inputs form opened at its key (that slice
/// owns those facts, ADR-0009). `null` when the target is unknown to this
/// app version or needs a regime the server did not name.
String? annualInputLocation(
  AnnualMissingInput input, {
  required String taxpayerId,
  required int fiscalYear,
}) {
  final regime = input.regimeCode;
  return switch (input.target) {
    AnnualInputTarget.salaryStatements => AnnualRoutes.salaryStatements(
      taxpayerId,
      fiscalYear,
    ),
    AnnualInputTarget.personalDeductions => AnnualRoutes.personalDeductions(
      taxpayerId,
      fiscalYear,
    ),
    AnnualInputTarget.annualInputs => AnnualRoutes.annualInputs(
      taxpayerId,
      fiscalYear,
    ),
    AnnualInputTarget.filedPayments when regime != null =>
      AnnualRoutes.filedPayments(taxpayerId, fiscalYear, regime),
    AnnualInputTarget.taxInputs when regime != null =>
      TaxEngineRoutes.taxInputs(
        taxpayerId,
        regime,
        fiscalYear: fiscalYear,
        field: input.key,
      ),
    _ => null,
  };
}
