// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

@JsonEnum()
enum TaxCalendarEntryMovedReason {
  @JsonValue('weekend')
  weekend('weekend'),
  @JsonValue('holiday')
  holiday('holiday'),

  /// The name has been replaced because it contains a keyword. Original name: `null`.
  @JsonValue(null)
  valueNull(null),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const TaxCalendarEntryMovedReason(this.json);

  factory TaxCalendarEntryMovedReason.fromJson(String? json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<TaxCalendarEntryMovedReason> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
