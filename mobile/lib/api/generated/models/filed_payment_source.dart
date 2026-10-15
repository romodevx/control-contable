// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// Where a month's credited payment comes from (precedence filed → tax input → app estimate, OQ-A1). `none`: no payment for the month (outside the regime's window, or not closed yet in the current year).
@JsonEnum()
enum FiledPaymentSource {
  @JsonValue('filed')
  filed('filed'),
  @JsonValue('tax_input')
  taxInput('tax_input'),
  @JsonValue('estimated_by_app')
  estimatedByApp('estimated_by_app'),
  @JsonValue('none')
  none('none'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const FiledPaymentSource(this.json);

  factory FiledPaymentSource.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<FiledPaymentSource> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
