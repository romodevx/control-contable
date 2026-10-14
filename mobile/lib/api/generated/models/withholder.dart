// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// Who withholds: `counterparty` (the payer of an income), `you` (the taxpayer, a persona moral paying a supplier, who must remit it), `platform` (625, and platform sales of a persona moral) or `none`.
@JsonEnum()
enum Withholder {
  @JsonValue('counterparty')
  counterparty('counterparty'),
  @JsonValue('you')
  you('you'),
  @JsonValue('platform')
  platform('platform'),
  @JsonValue('none')
  none('none'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const Withholder(this.json);

  factory Withholder.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<Withholder> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
