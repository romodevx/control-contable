// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'decimal_rate.dart';
import 'money_cents.dart';

part 'tax_input_value.freezed.dart';
part 'tax_input_value.g.dart';

/// Exactly one of value_cents (kind=money) or value_rate (kind=rate).
@Freezed()
abstract class TaxInputValue with _$TaxInputValue {
  const factory TaxInputValue({
    required String key,
    @JsonKey(includeIfNull: false, name: 'value_cents') MoneyCents? valueCents,
    @JsonKey(includeIfNull: false, name: 'value_rate') DecimalRate? valueRate,
  }) = _TaxInputValue;

  factory TaxInputValue.fromJson(Map<String, Object?> json) =>
      _$TaxInputValueFromJson(json);
}
