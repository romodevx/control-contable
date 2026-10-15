// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// Append-only (like `ErrorCode`, not frozen): the single list of `Warning.code` values. Clients show the `message` of a code they do not know. Taxpayers (M2): `rfc_check_digit_mismatch` on `rfc` (D-RFC); `regime_combination_needs_cpa_review` on `regimes.N.regime_code` (create) or `regime_code` (add regime), the pair overlaps in time and its compatibility is not yet verified by a CPA, `message` carries the pair's `reason`. Incomes (M3): `duplicate_cfdi_uuid` on `cfdi_uuid` (another live record of the taxpayer has the same folio fiscal); `withholding_differs_from_rate` on `isr_withheld_cents` / `iva_withheld_cents` (the CFDI wins); `provisional_parameters` on `collected_on` / `invoiced_on` (the rates checked on that date come from a provisional parameter set: its fiscal year is not published yet, so the previous year's values apply, or the set is marked provisional; M3-04). Expenses (M4): `cash_over_threshold` on `payment_method`; `no_cfdi` on `has_cfdi`; `not_isr_deductible_in_regime` and `category_not_deductible` on `deductible`; `duplicate_cfdi_uuid` on `cfdi_uuid`; `iva_not_creditable_cash_payment` on `payment_method` (a cash payment over the limit claimed as deductible: the claim stands for ISR, its IVA is not creditable, LIVA 5 fr. II ⚠️; 0.6.0); `provisional_parameters` on `paid_on` (the payment date's fiscal year has no published tax parameters yet, so the IVA and cash rules used the previous year's; M4-05). Withholdings (MODULE-withholdings §6, ADR-0008; 0.10.0): `withholding_cpa_unverified` on `withholding` (a ⚠️ rule a CPA has not confirmed matched); `foreign_counterparty_not_supported` on `counterparty_kind` (incomes) or `supplier_kind` (expenses); `supplier_regime_unknown` on `supplier_regime` (derived as 612); `deduction_requires_remittance` on `withholding` (an expense with a withholding above 0 is deductible only if it is remitted by the 17th of the next month). `withholding_differs_from_rate` is **retired**: WH-06 stops sending it, and it stays in the list because the list is append-only.
@JsonEnum()
enum WarningCode {
  @JsonValue('rfc_check_digit_mismatch')
  rfcCheckDigitMismatch('rfc_check_digit_mismatch'),
  @JsonValue('regime_combination_needs_cpa_review')
  regimeCombinationNeedsCpaReview('regime_combination_needs_cpa_review'),
  @JsonValue('duplicate_cfdi_uuid')
  duplicateCfdiUuid('duplicate_cfdi_uuid'),
  @JsonValue('withholding_differs_from_rate')
  withholdingDiffersFromRate('withholding_differs_from_rate'),
  @JsonValue('cash_over_threshold')
  cashOverThreshold('cash_over_threshold'),
  @JsonValue('no_cfdi')
  noCfdi('no_cfdi'),
  @JsonValue('not_isr_deductible_in_regime')
  notIsrDeductibleInRegime('not_isr_deductible_in_regime'),
  @JsonValue('category_not_deductible')
  categoryNotDeductible('category_not_deductible'),
  @JsonValue('iva_not_creditable_cash_payment')
  ivaNotCreditableCashPayment('iva_not_creditable_cash_payment'),
  @JsonValue('provisional_parameters')
  provisionalParameters('provisional_parameters'),
  @JsonValue('withholding_cpa_unverified')
  withholdingCpaUnverified('withholding_cpa_unverified'),
  @JsonValue('foreign_counterparty_not_supported')
  foreignCounterpartyNotSupported('foreign_counterparty_not_supported'),
  @JsonValue('supplier_regime_unknown')
  supplierRegimeUnknown('supplier_regime_unknown'),
  @JsonValue('deduction_requires_remittance')
  deductionRequiresRemittance('deduction_requires_remittance'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const WarningCode(this.json);

  factory WarningCode.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<WarningCode> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
