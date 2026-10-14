// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'decimal_rate.dart';
import 'non_negative_cents.dart';

part 'tax_input_value.freezed.dart';
part 'tax_input_value.g.dart';

/// One stored value. Exactly the field of the key's `kind` is set (`value_cents` for `money` and `money_monthly`, `value_rate` for `rate`, `value_bool` for `boolean`); the others are absent or `null`. `month` (1-12) is set only for `money_monthly`. Responses send every field, `null` when not set.
@Freezed()
abstract class TaxInputValue with _$TaxInputValue {
  const factory TaxInputValue({
    required String key,

    /// Month of the fiscal year, only for `money_monthly`.
    @JsonKey(includeIfNull: false) int? month,
    @JsonKey(includeIfNull: false, name: 'value_cents')
    NonNegativeCents? valueCents,
    @JsonKey(includeIfNull: false, name: 'value_rate') DecimalRate? valueRate,
    @JsonKey(includeIfNull: false, name: 'value_bool') bool? valueBool,
  }) = _TaxInputValue;

  factory TaxInputValue.fromJson(Map<String, Object?> json) =>
      _$TaxInputValueFromJson(json);
}
