// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'non_negative_cents.dart';
import 'regime_code.dart';

part 'annual_regime_tax_input.freezed.dart';
part 'annual_regime_tax_input.g.dart';

/// Read-only mirror of one annual-relevant M5 tax input (`money` kind) of a regime held in the year; edited with `putTaxInputs`. `value_cents: null` = not entered.
@Freezed()
abstract class AnnualRegimeTaxInput with _$AnnualRegimeTaxInput {
  const factory AnnualRegimeTaxInput({
    @JsonKey(name: 'regime_code') required RegimeCode regimeCode,
    required String key,
    required String label,
    @JsonKey(includeIfNull: true, name: 'value_cents')
    required NonNegativeCents? valueCents,
  }) = _AnnualRegimeTaxInput;

  factory AnnualRegimeTaxInput.fromJson(Map<String, Object?> json) =>
      _$AnnualRegimeTaxInputFromJson(json);
}
