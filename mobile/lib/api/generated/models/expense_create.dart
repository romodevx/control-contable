// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'iva_rate.dart';
import 'non_negative_cents.dart';
import 'payment_method.dart';
import 'regime_code.dart';
import 'uuid.dart';

part 'expense_create.freezed.dart';
part 'expense_create.g.dart';

@Freezed()
abstract class ExpenseCreate with _$ExpenseCreate {
  const factory ExpenseCreate({
    @JsonKey(name: 'regime_code') RegimeCode? regimeCode,
    @JsonKey(name: 'category_code') String? categoryCode,
    String? description,

    /// Subtotal before IVA.
    @JsonKey(name: 'amount_cents') NonNegativeCents? amountCents,
    @JsonKey(name: 'iva_rate') IvaRate? ivaRate,
    @JsonKey(name: 'iva_cents') NonNegativeCents? ivaCents,
    @JsonKey(name: 'payment_method') PaymentMethod? paymentMethod,
    @JsonKey(name: 'has_cfdi') bool? hasCfdi,
    @JsonKey(name: 'paid_on') DateTime? paidOn,

    /// ISR the taxpayer (typically a PM) withheld from the supplier; must be remitted.
    @JsonKey(name: 'isr_withheld_cents') NonNegativeCents? isrWithheldCents,

    /// IVA the taxpayer withheld from the supplier (LIVA 1-A); remitted in full, creditable when paid.
    @JsonKey(name: 'iva_withheld_cents') NonNegativeCents? ivaWithheldCents,
    Uuid? id,
  }) = _ExpenseCreate;

  factory ExpenseCreate.fromJson(Map<String, Object?> json) =>
      _$ExpenseCreateFromJson(json);
}
