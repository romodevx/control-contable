// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

@JsonEnum()
enum AnnualBreakdownLineOperation {
  @JsonValue('input')
  input('input'),
  @JsonValue('add')
  add('add'),
  @JsonValue('subtract')
  subtract('subtract'),
  @JsonValue('multiply')
  multiply('multiply'),
  @JsonValue('lookup')
  lookup('lookup'),
  @JsonValue('result')
  result('result'),
  @JsonValue('rounding_adjustment')
  roundingAdjustment('rounding_adjustment'),
  @JsonValue('info')
  info('info'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const AnnualBreakdownLineOperation(this.json);

  factory AnnualBreakdownLineOperation.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<AnnualBreakdownLineOperation> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
