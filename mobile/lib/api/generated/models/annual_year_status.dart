// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// `in_progress` when the fiscal year is the current one (America/Mexico_City): figures use the records through `data_through`, never extrapolated (OQ-A3).
@JsonEnum()
enum AnnualYearStatus {
  @JsonValue('closed')
  closed('closed'),
  @JsonValue('in_progress')
  inProgress('in_progress'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const AnnualYearStatus(this.json);

  factory AnnualYearStatus.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<AnnualYearStatus> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
