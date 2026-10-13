// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// rate_8 is the northern/southern border stimulus rate. Validity is date-ranged.
/// tax-parameter data (in force until 2026-12-31 per DOF 31-12-2025); a rate not valid.
/// on the record date is rejected with 422 business_rule_violation.
///
@JsonEnum()
enum IvaRate {
  @JsonValue('rate_16')
  rate16('rate_16'),
  @JsonValue('rate_8')
  rate8('rate_8'),
  @JsonValue('rate_0')
  rate0('rate_0'),
  @JsonValue('exempt')
  exempt('exempt'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const IvaRate(this.json);

  factory IvaRate.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<IvaRate> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
