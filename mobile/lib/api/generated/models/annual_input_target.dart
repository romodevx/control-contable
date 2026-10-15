// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// The screen that completes a missing input: `listSalaryStatements`, `listPersonalDeductions`, `getAnnualInputs`, `getFiledPayments` (for `regime_code`) or the M5 `getTaxInputs` (for `regime_code`).
@JsonEnum()
enum AnnualInputTarget {
  @JsonValue('salary_statements')
  salaryStatements('salary_statements'),
  @JsonValue('personal_deductions')
  personalDeductions('personal_deductions'),
  @JsonValue('annual_inputs')
  annualInputs('annual_inputs'),
  @JsonValue('filed_payments')
  filedPayments('filed_payments'),
  @JsonValue('tax_inputs')
  taxInputs('tax_inputs'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const AnnualInputTarget(this.json);

  factory AnnualInputTarget.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<AnnualInputTarget> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
