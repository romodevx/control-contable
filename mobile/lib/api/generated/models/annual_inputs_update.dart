// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'non_negative_cents.dart';

part 'annual_inputs_update.freezed.dart';
part 'annual_inputs_update.g.dart';

/// The complete set of per-year facts (replace, not merge).
@Freezed()
abstract class AnnualInputsUpdate with _$AnnualInputsUpdate {
  const factory AnnualInputsUpdate({
    @JsonKey(includeIfNull: true, name: 'prior_year_accumulable_income_cents')
    required NonNegativeCents? priorYearAccumulableIncomeCents,
    @JsonKey(name: 'other_exempt_income_cents')
    required NonNegativeCents otherExemptIncomeCents,
    @JsonKey(name: 'told_employer_will_file')
    required bool toldEmployerWillFile,
  }) = _AnnualInputsUpdate;

  factory AnnualInputsUpdate.fromJson(Map<String, Object?> json) =>
      _$AnnualInputsUpdateFromJson(json);
}
