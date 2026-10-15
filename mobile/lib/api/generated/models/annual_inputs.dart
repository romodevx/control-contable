// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'annual_regime_tax_input.dart';
import 'fiscal_year.dart';
import 'non_negative_cents.dart';

part 'annual_inputs.freezed.dart';
part 'annual_inputs.g.dart';

/// Per-year annual facts (MODULE-annual §2.6). `prior_year_accumulable_income_cents`: the previous year's acumulables before personal deductions, for the 7 % donation cap (`null` = unknown, OQ-A5). `other_exempt_income_cents`: exempt income outside the records, for the 15 % cap's total income. `told_employer_will_file`: the employee told the employer they will file (obligation test). `regime_tax_inputs`: read-only, ordered by regime then the catalog order of keys. `updated_at: null` = never saved (defaults).
@Freezed()
abstract class AnnualInputs with _$AnnualInputs {
  const factory AnnualInputs({
    @JsonKey(name: 'fiscal_year') required FiscalYear fiscalYear,
    @JsonKey(includeIfNull: true, name: 'prior_year_accumulable_income_cents')
    required NonNegativeCents? priorYearAccumulableIncomeCents,
    @JsonKey(name: 'other_exempt_income_cents')
    required NonNegativeCents otherExemptIncomeCents,
    @JsonKey(name: 'told_employer_will_file')
    required bool toldEmployerWillFile,
    @JsonKey(name: 'regime_tax_inputs')
    required List<AnnualRegimeTaxInput> regimeTaxInputs,
    @JsonKey(includeIfNull: true, name: 'updated_at')
    required DateTime? updatedAt,
  }) = _AnnualInputs;

  factory AnnualInputs.fromJson(Map<String, Object?> json) =>
      _$AnnualInputsFromJson(json);
}
