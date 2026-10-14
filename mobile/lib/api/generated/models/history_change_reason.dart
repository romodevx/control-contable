// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// Why a viewed month's displayed figures changed: `engine` when the engine version or the parameter set differs from the one viewed (it wins when both causes apply), `records` otherwise (a record, tax input or regime changed). es-MX copy: `engine` → "Se actualizó el cálculo desde que lo viste"; `records` → "Cambió desde que lo viste: hubo cambios en tus registros o datos fiscales".
@JsonEnum()
enum HistoryChangeReason {
  @JsonValue('records')
  records('records'),
  @JsonValue('engine')
  engine('engine'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const HistoryChangeReason(this.json);

  factory HistoryChangeReason.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<HistoryChangeReason> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
