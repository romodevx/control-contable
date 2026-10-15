// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// Whether the return must be filed (MODULE-annual §2.4): `pf_art152` is `obligated`, `optional` (salary only within the limits: filing is optional, a refund is the usual reason) or `not_obligated`; `pf_resico_check` is `relieved` (rule 3.13.7); PM sections are always `obligated`.
@JsonEnum()
enum AnnualObligation {
  @JsonValue('obligated')
  obligated('obligated'),
  @JsonValue('optional')
  optional('optional'),
  @JsonValue('not_obligated')
  notObligated('not_obligated'),
  @JsonValue('relieved')
  relieved('relieved'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const AnnualObligation(this.json);

  factory AnnualObligation.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<AnnualObligation> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
