// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'annual_input_target.dart';
import 'period.dart';
import 'regime_code.dart';

part 'annual_missing_input.freezed.dart';
part 'annual_missing_input.g.dart';

/// A required input the section does not have (MODULE-annual §2.6): `salary_statement` (the taxpayer holds 605 and has no constancia; salary counted as 0), `prior_year_accumulable_income` (donations exist; they are excluded), `filed_payment` (months whose default payment came from an `incomplete` monthly estimate; `months` lists them), or an M5 tax-input key (`target: tax_inputs`). `regime_code` is set for `filed_payments` and `tax_inputs`, `null` otherwise; `months` is empty unless the key is monthly.
@Freezed()
abstract class AnnualMissingInput with _$AnnualMissingInput {
  const factory AnnualMissingInput({
    required String key,

    /// Spanish label for the CTA.
    required String label,
    required AnnualInputTarget target,
    @JsonKey(includeIfNull: true, name: 'regime_code')
    required RegimeCode? regimeCode,
    required List<Period> months,
  }) = _AnnualMissingInput;

  factory AnnualMissingInput.fromJson(Map<String, Object?> json) =>
      _$AnnualMissingInputFromJson(json);
}
