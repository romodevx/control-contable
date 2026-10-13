// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'iva_components.dart';
import 'non_negative_cents.dart';
import 'parameter_set_ref.dart';
import 'regime_code.dart';
import 'regime_estimate_status.dart';
import 'tax_result.dart';
import 'treatment.dart';

part 'regime_estimate.freezed.dart';
part 'regime_estimate.g.dart';

@Freezed()
abstract class RegimeEstimate with _$RegimeEstimate {
  const factory RegimeEstimate({
    @JsonKey(name: 'regime_code') required RegimeCode regimeCode,
    @JsonKey(name: 'regime_name') required String regimeName,
    required Treatment treatment,
    required RegimeEstimateStatus status,
    @JsonKey(name: 'parameter_set') required ParameterSetRef parameterSet,
    @JsonKey(includeIfNull: false, name: 'missing_inputs')
    List<String>? missingInputs,
    @JsonKey(includeIfNull: false) String? explanation,
    @JsonKey(includeIfNull: false) TaxResult? isr,
    @JsonKey(includeIfNull: false, name: 'iva_components')
    IvaComponents? ivaComponents,

    /// ISR/IVA the taxpayer withheld from suppliers in this regime.
    @JsonKey(includeIfNull: false, name: 'withheld_to_remit_cents')
    NonNegativeCents? withheldToRemitCents,
  }) = _RegimeEstimate;

  factory RegimeEstimate.fromJson(Map<String, Object?> json) =>
      _$RegimeEstimateFromJson(json);
}
