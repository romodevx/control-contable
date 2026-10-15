// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// Precedence `informational` > `incomplete` > `complete`. `incomplete`: a required input is missing and the section is computed with its documented fallback ("Incompleto" chip and CTA). `informational`: `pf_resico_check` and `explanation`; no payable or refundable balance.
@JsonEnum()
enum AnnualSectionStatus {
  @JsonValue('complete')
  complete('complete'),
  @JsonValue('incomplete')
  incomplete('incomplete'),
  @JsonValue('informational')
  informational('informational'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const AnnualSectionStatus(this.json);

  factory AnnualSectionStatus.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<AnnualSectionStatus> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
