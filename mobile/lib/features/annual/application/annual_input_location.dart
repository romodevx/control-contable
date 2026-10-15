import 'package:control_contable/api/generated/models/annual_input_target.dart';
import 'package:control_contable/api/generated/models/annual_missing_input.dart';
import 'package:control_contable/features/tax_engine/tax_engine_routes.dart';

/// Where [input] is completed for [taxpayerId] in [fiscalYear]: the
/// monthly tax-inputs form opened at its key (that slice owns those facts,
/// ADR-0009). `null` when this app version has no screen for it: the
/// annual input screens ship with ANN-19..21, and until then the input is
/// named without a link, never sent to a dead end. Also `null` when the
/// target is unknown or needs a regime the server did not name.
String? annualInputLocation(
  AnnualMissingInput input, {
  required String taxpayerId,
  required int fiscalYear,
}) {
  final regime = input.regimeCode;
  return switch (input.target) {
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
