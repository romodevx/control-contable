// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

@JsonEnum()
enum DisclaimerCode {
  @JsonValue('not_tax_advice')
  notTaxAdvice('not_tax_advice'),
  @JsonValue('incomplete_inputs')
  incompleteInputs('incomplete_inputs'),
  @JsonValue('provisional_parameters')
  provisionalParameters('provisional_parameters'),
  @JsonValue('informational_regime')
  informationalRegime('informational_regime'),
  @JsonValue('deductibility_claimed')
  deductibilityClaimed('deductibility_claimed'),
  @JsonValue('no_coefficient_first_year')
  noCoefficientFirstYear('no_coefficient_first_year'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const DisclaimerCode(this.json);

  factory DisclaimerCode.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<DisclaimerCode> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
