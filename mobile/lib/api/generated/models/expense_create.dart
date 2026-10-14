// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'business_date.dart';
import 'cfdi_uuid.dart';
import 'iva_rate.dart';
import 'non_negative_cents.dart';
import 'payment_method.dart';
import 'positive_cents.dart';
import 'record_description.dart';
import 'regime_code.dart';
import 'uuid.dart';

part 'expense_create.freezed.dart';
part 'expense_create.g.dart';

/// `user_id` comes from the session and `taxpayer_id` from the path, never from the body.
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

    /// ISR the taxpayer (a persona moral) withheld from the supplier (default 0); at most `amount_cents`.
    @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')
    NonNegativeCents? isrWithheldCents,

    /// IVA the taxpayer (a persona moral) withheld from the supplier (LIVA 1-A, default 0); at most `iva_cents`.
    @JsonKey(includeIfNull: false, name: 'iva_withheld_cents')
    NonNegativeCents? ivaWithheldCents,
    @JsonKey(includeIfNull: false) RecordDescription? description,
  }) = _ExpenseCreate;

  factory ExpenseCreate.fromJson(Map<String, Object?> json) =>
      _$ExpenseCreateFromJson(json);
}
