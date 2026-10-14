// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// `strong` is the non-dismissible amber banner above the number (`cpa_review_pending`); `warning` and `info` are shown under the estimate.
@JsonEnum()
enum DisclaimerSeverity {
  @JsonValue('info')
  info('info'),
  @JsonValue('warning')
  warning('warning'),
  @JsonValue('strong')
  strong('strong'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const DisclaimerSeverity(this.json);

  factory DisclaimerSeverity.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<DisclaimerSeverity> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
