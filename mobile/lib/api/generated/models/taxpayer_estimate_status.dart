// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// Taxpayer-level status of one month's estimate, derived server-side from its regimes with the M5 precedence: `informational` (every regime informational) > `incomplete` (any regime incomplete) > `partial` (any partial) > `complete`; `none` when the taxpayer holds no regime that month (`regimes: []`).
@JsonEnum()
enum TaxpayerEstimateStatus {
  @JsonValue('complete')
  complete('complete'),
  @JsonValue('incomplete')
  incomplete('incomplete'),
  @JsonValue('partial')
  partial('partial'),
  @JsonValue('informational')
  informational('informational'),
  @JsonValue('none')
  none('none'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const TaxpayerEstimateStatus(this.json);

  factory TaxpayerEstimateStatus.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<TaxpayerEstimateStatus> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
