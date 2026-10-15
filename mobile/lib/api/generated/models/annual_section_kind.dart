// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// `pf_art152`: one aggregated LISR 152 computation of a persona física's 605, 606, 612 and 625 non-definitive income. `pf_resico_check`: RESICO PF 113-F informative check, never a balance. `pm_601`, `pm_resico`: per regime (LISR 9). `explanation`: any other regime, text and deadline only.
@JsonEnum()
enum AnnualSectionKind {
  @JsonValue('pf_art152')
  pfArt152('pf_art152'),
  @JsonValue('pf_resico_check')
  pfResicoCheck('pf_resico_check'),
  @JsonValue('pm_601')
  pm601('pm_601'),
  @JsonValue('pm_resico')
  pmResico('pm_resico'),
  @JsonValue('explanation')
  explanation('explanation'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const AnnualSectionKind(this.json);

  factory AnnualSectionKind.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<AnnualSectionKind> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
