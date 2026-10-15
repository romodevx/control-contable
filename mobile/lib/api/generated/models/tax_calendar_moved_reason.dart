// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// Why the nominal date (day 17) is not the due date: `weekend` when day 17 is a Saturday or Sunday, `holiday` when it is a holiday that moves due dates. The days skipped after it may be of either kind.
@JsonEnum()
enum TaxCalendarMovedReason {
  @JsonValue('weekend')
  weekend('weekend'),
  @JsonValue('holiday')
  holiday('holiday'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const TaxCalendarMovedReason(this.json);

  factory TaxCalendarMovedReason.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<TaxCalendarMovedReason> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
