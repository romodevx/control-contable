// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// "¿Quién te vendió o te dio el servicio?" (MODULE-withholdings §2.2 E1), only when the taxpayer is a persona moral. `foreign` derives zero and warns `foreign_counterparty_not_supported`.
@JsonEnum()
enum SupplierKind {
  @JsonValue('pm')
  pm('pm'),
  @JsonValue('pf_with_activity')
  pfWithActivity('pf_with_activity'),
  @JsonValue('foreign')
  foreign('foreign'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const SupplierKind(this.json);

  factory SupplierKind.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<SupplierKind> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
