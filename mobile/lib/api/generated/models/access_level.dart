// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// The caller's access (P0-16, PRD §7): `trial` (running trial, capped records), `pro` (store plan), `comp` (granted), `read_only` (trial over and no plan: reads, deletes and restores stay, creates and edits answer `402`).
@JsonEnum()
enum AccessLevel {
  @JsonValue('trial')
  trial('trial'),
  @JsonValue('pro')
  pro('pro'),
  @JsonValue('comp')
  comp('comp'),
  @JsonValue('read_only')
  readOnly('read_only'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const AccessLevel(this.json);

  factory AccessLevel.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<AccessLevel> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
