// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// "¿A quién le facturaste?" (MODULE-withholdings §2.1 Q1): `pm` a persona moral, `pf_with_activity` a persona física with a business or profession, `pf_no_activity_or_public` a persona física without activity or the general public, `gov_federal` the federal government or a federal body, `gov_state_municipal` a state or municipal government, `foreign` someone abroad (derives zero and warns `foreign_counterparty_not_supported`). The payer's regime is never asked: it cannot change the result (OQ-W1).
@JsonEnum()
enum CounterpartyKind {
  @JsonValue('pm')
  pm('pm'),
  @JsonValue('pf_with_activity')
  pfWithActivity('pf_with_activity'),
  @JsonValue('pf_no_activity_or_public')
  pfNoActivityOrPublic('pf_no_activity_or_public'),
  @JsonValue('gov_federal')
  govFederal('gov_federal'),
  @JsonValue('gov_state_municipal')
  govStateMunicipal('gov_state_municipal'),
  @JsonValue('foreign')
  foreign('foreign'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const CounterpartyKind(this.json);

  factory CounterpartyKind.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<CounterpartyKind> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
