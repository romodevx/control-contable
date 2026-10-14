// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// Regime of a persona física supplier as printed on its CFDI (MODULE-withholdings §2.2 E2), only with `supplier_kind: pf_with_activity`. `unknown` ("No sé") derives as 612 and warns `supplier_regime_unknown` (612 fees withhold 10 % ISR, RESICO 1.25 %).
@JsonEnum()
enum SupplierRegime {
  @JsonValue('612')
  value612('612'),
  @JsonValue('606')
  value606('606'),
  @JsonValue('626')
  value626('626'),
  @JsonValue('625')
  value625('625'),
  @JsonValue('621')
  value621('621'),
  @JsonValue('unknown')
  unknown('unknown'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const SupplierRegime(this.json);

  factory SupplierRegime.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<SupplierRegime> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
