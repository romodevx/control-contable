// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// `donation` only: an authorized donee (7 % cap) or a government entity (4 % within the 7 %).
@JsonEnum()
enum DoneeType {
  @JsonValue('authorized')
  authorized('authorized'),
  @JsonValue('government')
  government('government'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const DoneeType(this.json);

  factory DoneeType.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<DoneeType> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
