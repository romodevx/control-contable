// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// `tuition` only; the cap of the highest level the student had in the year applies.
@JsonEnum()
enum TuitionLevel {
  @JsonValue('preescolar')
  preescolar('preescolar'),
  @JsonValue('primaria')
  primaria('primaria'),
  @JsonValue('secundaria')
  secundaria('secundaria'),
  @JsonValue('profesional_tecnico')
  profesionalTecnico('profesional_tecnico'),
  @JsonValue('bachillerato')
  bachillerato('bachillerato'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const TuitionLevel(this.json);

  factory TuitionLevel.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<TuitionLevel> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
