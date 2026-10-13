// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'person_type.dart';
import 'regime_code.dart';
import 'regime_option_def.dart';
import 'treatment.dart';

part 'regime.freezed.dart';
part 'regime.g.dart';

/// Catalog key is (code, person_type); e.g. 626 exists as RESICO PF and RESICO PM.
@Freezed()
abstract class Regime with _$Regime {
  const factory Regime({
    required RegimeCode code,
    required String name,
    @JsonKey(name: 'person_type') required PersonType personType,
    required Treatment treatment,
    @JsonKey(name: 'option_defs') required List<RegimeOptionDef> optionDefs,

    /// Tax-input keys the strategy needs (see TaxInputs).
    @JsonKey(name: 'required_inputs') required List<String> requiredInputs,

    /// Plain-language reason for partial/informational treatment.
    @JsonKey(includeIfNull: false) String? explanation,
  }) = _Regime;

  factory Regime.fromJson(Map<String, Object?> json) => _$RegimeFromJson(json);
}
