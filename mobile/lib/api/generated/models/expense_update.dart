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

part 'expense_update.freezed.dart';
part 'expense_update.g.dart';

/// Merge-patch on top-level fields: absent = unchanged, `null` clears `cfdi_uuid`, `description` or the `deductible` override. `id` is immutable and not accepted.
@Freezed()
abstract class ExpenseUpdate with _$ExpenseUpdate {
  const factory ExpenseUpdate({
    @JsonKey(includeIfNull: false, name: 'regime_code') RegimeCode? regimeCode,
    @JsonKey(includeIfNull: false, name: 'category_code') String? categoryCode,
    @JsonKey(includeIfNull: false, name: 'amount_cents')
    PositiveCents? amountCents,
    @JsonKey(includeIfNull: false, name: 'iva_rate') IvaRate? ivaRate,
    @JsonKey(includeIfNull: false, name: 'iva_cents')
    NonNegativeCents? ivaCents,
    @JsonKey(includeIfNull: false, name: 'payment_method')
    PaymentMethod? paymentMethod,
    @JsonKey(includeIfNull: false, name: 'paid_on') BusinessDate? paidOn,
    @JsonKey(includeIfNull: false, name: 'has_cfdi') bool? hasCfdi,
    @JsonKey(includeIfNull: false, name: 'cfdi_uuid') CfdiUuid? cfdiUuid,
    @JsonKey(includeIfNull: false) bool? deductible,
    @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')
    NonNegativeCents? isrWithheldCents,
    @JsonKey(includeIfNull: false, name: 'iva_withheld_cents')
    NonNegativeCents? ivaWithheldCents,
    @JsonKey(includeIfNull: false) RecordDescription? description,
  }) = _ExpenseUpdate;

  factory ExpenseUpdate.fromJson(Map<String, Object?> json) =>
      _$ExpenseUpdateFromJson(json);
}
