// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// The `AnnualSectionKind` values that have a computation trace.
@JsonEnum()
enum AnnualComputedSectionKind {
  @JsonValue('pf_art152')
  pfArt152('pf_art152'),
  @JsonValue('pf_resico_check')
  pfResicoCheck('pf_resico_check'),
  @JsonValue('pm_601')
  pm601('pm_601'),
  @JsonValue('pm_resico')
  pmResico('pm_resico'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const AnnualComputedSectionKind(this.json);

  factory AnnualComputedSectionKind.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<AnnualComputedSectionKind> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
