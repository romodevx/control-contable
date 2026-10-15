// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'business_date.dart';
import 'cfdi_uuid.dart';
import 'deductible_source.dart';
import 'iva_rate.dart';
import 'legend.dart';
import 'non_negative_cents.dart';
import 'operation_kind.dart';
import 'payment_method.dart';
import 'positive_cents.dart';
import 'record_description.dart';
import 'regime_code.dart';
import 'supplier_kind.dart';
import 'supplier_regime.dart';
import 'uuid.dart';
import 'warning.dart';
import 'withholding.dart';

part 'expense.freezed.dart';
part 'expense.g.dart';

/// `withholding` is the derived withholding (read-only, ADR-0008): all zero with `withholder: none` for a persona física taxpayer; sent for every expense written since 0.11.3 (WH-07), absent only on a legacy row until the WH-08 backfill derives it.
@Freezed()
abstract class Expense with _$Expense {
  const factory Expense({
    required Uuid id,
    @JsonKey(name: 'taxpayer_id') required Uuid taxpayerId,
    @JsonKey(name: 'regime_code') required RegimeCode regimeCode,
    @JsonKey(name: 'category_code') required String categoryCode,
    @JsonKey(name: 'amount_cents') required PositiveCents amountCents,
    @JsonKey(name: 'iva_rate') required IvaRate ivaRate,
    @JsonKey(name: 'iva_cents') required NonNegativeCents ivaCents,
    @JsonKey(name: 'payment_method') required PaymentMethod paymentMethod,
    @JsonKey(name: 'paid_on') required BusinessDate paidOn,
    @JsonKey(name: 'has_cfdi') required bool hasCfdi,
    @JsonKey(includeIfNull: true, name: 'cfdi_uuid')
    required CfdiUuid? cfdiUuid,

    /// Read-only mirror of `withholding.isr_cents` (kept for list totals and the estimate).
    @JsonKey(name: 'isr_withheld_cents')
    required NonNegativeCents isrWithheldCents,

    /// Read-only mirror of `withholding.iva_cents` (kept for list totals and the estimate).
    @JsonKey(name: 'iva_withheld_cents')
    required NonNegativeCents ivaWithheldCents,
    required RecordDescription description,

    /// Effective value ("the user claims it lowers ISR"): the override if any, else the default.
    required bool deductible,
    @JsonKey(name: 'deductible_source')
    required DeductibleSource deductibleSource,

    /// Non-blocking notices computed on writes (create, update, restore), never stored; `[]` on reads. Codes (append-only): `cash_over_threshold` (`payment_method`), `no_cfdi` (`has_cfdi`), `not_isr_deductible_in_regime` and `category_not_deductible` (`deductible`), `duplicate_cfdi_uuid` (`cfdi_uuid`), `iva_not_creditable_cash_payment` (`payment_method`), `provisional_parameters` (`paid_on`), `deduction_requires_remittance` (`withholding`), `withholding_cpa_unverified` (`withholding`), `supplier_regime_unknown` (`supplier_regime`), `foreign_counterparty_not_supported` (`supplier_kind`).
    required List<Warning> warnings,
    required Legend legend,
    @JsonKey(name: 'created_at') required DateTime createdAt,
    @JsonKey(name: 'updated_at') required DateTime updatedAt,

    /// `null` for a persona física taxpayer. Always sent (nullable) since 0.11.3.
    @JsonKey(includeIfNull: false, name: 'supplier_kind')
    SupplierKind? supplierKind,

    /// `null` unless `supplier_kind: pf_with_activity`. Always sent (nullable) since 0.11.3.
    @JsonKey(includeIfNull: false, name: 'supplier_regime')
    SupplierRegime? supplierRegime,

    /// `null` for a persona física taxpayer. Always sent (nullable) since 0.11.3.
    @JsonKey(includeIfNull: false, name: 'operation_kind')
    OperationKind? operationKind,

    /// Always sent since 0.11.3.
    @JsonKey(includeIfNull: false, name: 'residential_rent')
    bool? residentialRent,
    @JsonKey(includeIfNull: false) Withholding? withholding,
  }) = _Expense;

  factory Expense.fromJson(Map<String, Object?> json) =>
      _$ExpenseFromJson(json);
}
