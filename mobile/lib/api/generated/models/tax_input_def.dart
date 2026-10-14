// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'tax_input_def_kind.dart';

part 'tax_input_def.freezed.dart';
part 'tax_input_def.g.dart';

/// One per-fiscal-year tax input of a regime (`regimes.required_inputs`, catalog data). `kind`: `money` (one amount for the year, centavos `>= 0`), `money_monthly` (one amount per month 1-12, e.g. the prior provisional payments, OQ-T1), `rate` (decimal string in `[0, 1]`) or `boolean`. `required`: the key can be required in some period of the year; whether it is missing for a period is `RegimeEstimate.missing_inputs`.
@Freezed()
abstract class TaxInputDef with _$TaxInputDef {
  const factory TaxInputDef({
    required String key,

    /// Spanish label shown in the tax-inputs form.
    required String label,
    required TaxInputDefKind kind,

    /// The name has been replaced because it contains a keyword. Original name: `required`.
    @JsonKey(name: 'required') required bool requiredValue,
  }) = _TaxInputDef;

  factory TaxInputDef.fromJson(Map<String, Object?> json) =>
      _$TaxInputDefFromJson(json);
}
