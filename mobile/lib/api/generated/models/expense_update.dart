// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'iva_rate.dart';
import 'non_negative_cents.dart';
import 'payment_method.dart';
import 'regime_code.dart';

part 'expense_update.freezed.dart';
part 'expense_update.g.dart';

@Freezed()
abstract class ExpenseUpdate with _$ExpenseUpdate {
  const factory ExpenseUpdate({
    @JsonKey(includeIfNull: false, name: 'regime_code') RegimeCode? regimeCode,
    @JsonKey(includeIfNull: false, name: 'category_code') String? categoryCode,
    @JsonKey(includeIfNull: false) String? description,

    /// Subtotal before IVA.
    @JsonKey(includeIfNull: false, name: 'amount_cents')
    NonNegativeCents? amountCents,
    @JsonKey(includeIfNull: false, name: 'iva_rate') IvaRate? ivaRate,
    @JsonKey(includeIfNull: false, name: 'iva_cents')
    NonNegativeCents? ivaCents,
    @JsonKey(includeIfNull: false, name: 'payment_method')
    PaymentMethod? paymentMethod,
    @JsonKey(includeIfNull: false, name: 'has_cfdi') bool? hasCfdi,
    @JsonKey(includeIfNull: false, name: 'paid_on') DateTime? paidOn,

    /// ISR the taxpayer (typically a PM) withheld from the supplier; must be remitted.
    @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')
    NonNegativeCents? isrWithheldCents,

    /// IVA the taxpayer withheld from the supplier (LIVA 1-A); remitted in full, creditable when paid.
    @JsonKey(includeIfNull: false, name: 'iva_withheld_cents')
    NonNegativeCents? ivaWithheldCents,
  }) = _ExpenseUpdate;

  factory ExpenseUpdate.fromJson(Map<String, Object?> json) =>
      _$ExpenseUpdateFromJson(json);
}
