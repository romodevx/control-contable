// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// Append-only (like `ErrorCode`): clients show the `text` of a code they do not know. `not_tax_advice` always; `incomplete_inputs` when a regime is `incomplete`; `provisional_parameters` when the period's fiscal year has no published parameters yet; `informational_regime` / `partial_regime` when a regime has that status; `deductibility_claimed` when a deductible expense entered the ISR base; `no_coefficient_first_year` for a 601 in its first fiscal year; `iva_mixed_activity` when a regime's creditable IVA was taken in full although it has exempt acts (`iva_fully_taxable = false`, or exempt collections in the month; LIVA 5 fr. V); `cpa_review_pending` when a line applies a rule a CPA has not validated; `resico_cap_exceeded` when a RESICO (626) regime's income collected in the year is above its cap (PF 3,500,000, LISR 113-E; PM 35,000,000, LISR 206): the taxpayer leaves RESICO (PF from the next month, LISR 113-E/113-I; PM from the next fiscal year, LISR 214), the month is still estimated under RESICO (MODULE-tax-engine §2.3, §2.9). Annual estimate (MODULE-annual §2.8, 0.11.0; `annual_estimate` is first there instead of `not_tax_advice`, and `incomplete_inputs` / `cpa_review_pending` are reused): `annual_estimate` always; `year_in_progress` (current fiscal year); `payments_estimated` (a provisional payment taken from the app's estimate); `deductions_as_entered` (a personal deduction applied); `resico_pf_definitive` (a RESICO PF check section); `informative_income` (PF total income over 500,000.00, LISR 150); `rmf_2027_pending` (fiscal year 2026); `inflation_adjustment_excluded` (PM sections); `no_annual_iva` always.
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
  @JsonValue('resico_cap_exceeded')
  resicoCapExceeded('resico_cap_exceeded'),
  @JsonValue('annual_estimate')
  annualEstimate('annual_estimate'),
  @JsonValue('year_in_progress')
  yearInProgress('year_in_progress'),
  @JsonValue('payments_estimated')
  paymentsEstimated('payments_estimated'),
  @JsonValue('deductions_as_entered')
  deductionsAsEntered('deductions_as_entered'),
  @JsonValue('resico_pf_definitive')
  resicoPfDefinitive('resico_pf_definitive'),
  @JsonValue('informative_income')
  informativeIncome('informative_income'),
  @JsonValue('rmf_2027_pending')
  rmf2027Pending('rmf_2027_pending'),
  @JsonValue('inflation_adjustment_excluded')
  inflationAdjustmentExcluded('inflation_adjustment_excluded'),
  @JsonValue('no_annual_iva')
  noAnnualIva('no_annual_iva'),

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
