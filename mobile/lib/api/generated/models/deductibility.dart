// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// The hint shown for a category in the taxpayer's regime (after options such as 606 `blind_35`).
@JsonEnum()
enum Deductibility {
  @JsonValue('deductible')
  deductible('deductible'),
  @JsonValue('partially_deductible')
  partiallyDeductible('partially_deductible'),
  @JsonValue('not_deductible')
  notDeductible('not_deductible'),
  @JsonValue('isr_not_applicable')
  isrNotApplicable('isr_not_applicable'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const Deductibility(this.json);

  factory Deductibility.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<Deductibility> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
