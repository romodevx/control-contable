// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// Append-only (like `ErrorCode`): clients show the `text` of a code they do not know. `not_tax_advice` always; `incomplete_inputs` when a regime is `incomplete`; `provisional_parameters` when the period's fiscal year has no published parameters yet; `informational_regime` / `partial_regime` when a regime has that status; `deductibility_claimed` when a deductible expense entered the ISR base; `no_coefficient_first_year` for a 601 in its first fiscal year; `iva_mixed_activity` when `iva_fully_taxable = false` and creditable IVA was taken in full; `cpa_review_pending` when a line applies a rule a CPA has not validated (MODULE-tax-engine §2.3, §2.9).
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
  @JsonValue('partial_regime')
  partialRegime('partial_regime'),
  @JsonValue('deductibility_claimed')
  deductibilityClaimed('deductibility_claimed'),
  @JsonValue('no_coefficient_first_year')
  noCoefficientFirstYear('no_coefficient_first_year'),
  @JsonValue('iva_mixed_activity')
  ivaMixedActivity('iva_mixed_activity'),
  @JsonValue('cpa_review_pending')
  cpaReviewPending('cpa_review_pending'),

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
