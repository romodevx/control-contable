// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

@JsonEnum()
enum ExpenseWarningCode {
  @JsonValue('cash_over_threshold')
  cashOverThreshold('cash_over_threshold'),
  @JsonValue('no_cfdi')
  noCfdi('no_cfdi'),
  @JsonValue('not_isr_deductible_in_regime')
  notIsrDeductibleInRegime('not_isr_deductible_in_regime'),
  @JsonValue('category_not_deductible')
  categoryNotDeductible('category_not_deductible'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const ExpenseWarningCode(this.json);

  factory ExpenseWarningCode.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<ExpenseWarningCode> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
