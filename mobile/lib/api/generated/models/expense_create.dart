// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'business_date.dart';
import 'cfdi_uuid.dart';
import 'iva_rate.dart';
import 'non_negative_cents.dart';
import 'operation_kind.dart';
import 'payment_method.dart';
import 'positive_cents.dart';
import 'record_description.dart';
import 'regime_code.dart';
import 'supplier_kind.dart';
import 'supplier_regime.dart';
import 'uuid.dart';

part 'expense_create.freezed.dart';
part 'expense_create.g.dart';

/// `user_id` comes from the session and `taxpayer_id` from the path, never from the body. Withholdings are **derived** (ADR-0008): a persona moral taxpayer describes its supplier (`supplier_kind`, required for it since 0.11.3; `supplier_regime`; `operation_kind`, defaulting to the category's; `residential_rent`) and the server derives what it must withhold and remit. A persona física never withholds from a domestic supplier: any supplier field answers `422` `supplier_not_applicable`. **0.11.3 (WH-07):** the deprecated `isr_withheld_cents` and `iva_withheld_cents` with any value answer `400 additionalProperties`; they leave the schema with WH-12, which regenerates the Dart client with the new form.
@Freezed()
abstract class ExpenseCreate with _$ExpenseCreate {
  const factory ExpenseCreate({
    @JsonKey(name: 'regime_code') required RegimeCode regimeCode,

    /// A code of `listExpenseCategories` (`unknown_category` otherwise).
    @JsonKey(name: 'category_code') required String categoryCode,

    /// Subtotal before IVA, MXN centavos.
    @JsonKey(name: 'amount_cents') required PositiveCents amountCents,
    @JsonKey(name: 'iva_rate') required IvaRate ivaRate,

    /// IVA paid; `0` for `rate_0`/`exempt`, else within 100 centavos of `amount_cents × rate`.
    @JsonKey(name: 'iva_cents') required NonNegativeCents ivaCents,
    @JsonKey(name: 'payment_method') required PaymentMethod paymentMethod,

    /// The only date (paid basis for every regime in scope); the booking date.
    @JsonKey(name: 'paid_on') required BusinessDate paidOn,
    @JsonKey(name: 'has_cfdi') required bool hasCfdi,

    /// Client-generated idempotency key (UUIDv7 recommended), once per form session.
    @JsonKey(includeIfNull: false) Uuid? id,

    /// Requires `has_cfdi: true` (`has_cfdi_mismatch`).
    @JsonKey(includeIfNull: false, name: 'cfdi_uuid') CfdiUuid? cfdiUuid,

    /// The user's override; omitted or `null` uses the default.
    @JsonKey(includeIfNull: false) bool? deductible,

    /// Persona moral taxpayers only; required for them since 0.11.3 (`400 required`).
    @JsonKey(includeIfNull: false, name: 'supplier_kind')
    SupplierKind? supplierKind,

    /// Only with `supplier_kind: pf_with_activity` (omitted → `unknown`); with any other kind it answers `422` `supplier_not_applicable`.
    @JsonKey(includeIfNull: false, name: 'supplier_regime')
    SupplierRegime? supplierRegime,

    /// Persona moral taxpayers only; omitted → the category's default operation.
    @JsonKey(includeIfNull: false, name: 'operation_kind')
    OperationKind? operationKind,

    /// Persona moral taxpayers only, `rent_real_estate` only (default `false`). `true` requires `iva_rate: exempt` and `iva_cents: 0`, else `422` `residential_rent_exempt` on `iva_rate`.
    @JsonKey(includeIfNull: false, name: 'residential_rent')
    bool? residentialRent,

    /// Retired by WH-07 (0.11.3, ADR-0008). Any value answers `400 additionalProperties`; the server derives it.
    @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')
    NonNegativeCents? isrWithheldCents,

    /// Retired by WH-07 (0.11.3, ADR-0008). Any value answers `400 additionalProperties`; the server derives it.
    @JsonKey(includeIfNull: false, name: 'iva_withheld_cents')
    NonNegativeCents? ivaWithheldCents,
    @JsonKey(includeIfNull: false) RecordDescription? description,
  }) = _ExpenseCreate;

  factory ExpenseCreate.fromJson(Map<String, Object?> json) =>
      _$ExpenseCreateFromJson(json);
}
