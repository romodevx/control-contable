// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// **Deprecated in 0.10.0** (ADR-0008): replaced by `CounterpartyKind` plus `via_platform`. Removed from requests, responses and `getIncomeDefaults` by WH-06 (the contract step, 0.10.x). Until then it is who paid the income (MODULE-incomes §2.6): persona física and público en general never withhold; a persona moral withholds from a física in 612/606/626; a platform withholds in 625 and from a persona moral.
@JsonEnum()
enum CounterpartyType {
  @JsonValue('persona_fisica')
  personaFisica('persona_fisica'),
  @JsonValue('persona_moral')
  personaMoral('persona_moral'),
  @JsonValue('publico_general')
  publicoGeneral('publico_general'),
  @JsonValue('plataforma')
  plataforma('plataforma'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const CounterpartyType(this.json);

  factory CounterpartyType.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<CounterpartyType> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
