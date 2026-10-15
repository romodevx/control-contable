// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// Shown in words ("Saldo a cargo", "Saldo a favor", "Sin saldo"), never by colour alone.
@JsonEnum()
enum AnnualBalanceKind {
  @JsonValue('a_cargo')
  aCargo('a_cargo'),
  @JsonValue('a_favor')
  aFavor('a_favor'),
  @JsonValue('cero')
  cero('cero'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const AnnualBalanceKind(this.json);

  factory AnnualBalanceKind.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<AnnualBalanceKind> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
