// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// Chosen first at onboarding and immutable; fixes the RFC length (13 física, 12 moral).
@JsonEnum()
enum PersonType {
  @JsonValue('fisica')
  fisica('fisica'),
  @JsonValue('moral')
  moral('moral'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const PersonType(this.json);

  factory PersonType.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<PersonType> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
