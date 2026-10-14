// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'person_type.dart';
import 'regime_code.dart';
import 'regime_incompatibility.dart';
import 'regime_option_def.dart';
import 'tax_input_def.dart';
import 'treatment.dart';

part 'regime.freezed.dart';
part 'regime.g.dart';

/// Catalog key is `(code, person_type)`; e.g. 626 exists as RESICO PF and RESICO PM. `treatment` is the support level. `incompatible_with` is symmetric (if A lists B, B lists A) and only applies to overlapping period windows; an absent pair is allowed.
@Freezed()
abstract class Regime with _$Regime {
  const factory Regime({
    required RegimeCode code,
    required String name,
    @JsonKey(name: 'person_type') required PersonType personType,
    required Treatment treatment,

    /// Plain-language explanation for a partial or informational regime; `null` for full.
    @JsonKey(includeIfNull: true) required String? explanation,
    @JsonKey(name: 'option_defs') required List<RegimeOptionDef> optionDefs,

    /// Per-fiscal-year tax inputs the estimate needs (`PUT …/tax-inputs`, M5); not options. Exactly the stored `regimes.required_inputs` (CHECK `regime_required_inputs_valid`).
    @JsonKey(name: 'required_inputs') required List<TaxInputDef> requiredInputs,
    @JsonKey(name: 'incompatible_with')
    required List<RegimeIncompatibility> incompatibleWith,
  }) = _Regime;

  factory Regime.fromJson(Map<String, Object?> json) => _$RegimeFromJson(json);
}
