// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// Anyone but `self` means the user attested that the dependant's income is below one UMA anual (not verified ⚠️).
@JsonEnum()
enum DeductionBeneficiary {
  @JsonValue('self')
  self('self'),
  @JsonValue('spouse')
  spouse('spouse'),
  @JsonValue('ascendant')
  ascendant('ascendant'),
  @JsonValue('descendant')
  descendant('descendant'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const DeductionBeneficiary(this.json);

  factory DeductionBeneficiary.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<DeductionBeneficiary> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
