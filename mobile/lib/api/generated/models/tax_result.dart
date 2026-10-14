// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'money_cents.dart';
import 'non_negative_cents.dart';
import 'tax_result_outcome.dart';

part 'tax_result.freezed.dart';
part 'tax_result.g.dart';

/// One tax result, rounded once at the end to the centavo, half away from zero (ADR-0002 §4). `in_favor` is a saldo a favor: never shown as a payment.
@Freezed()
abstract class TaxResult with _$TaxResult {
  const factory TaxResult({
    /// Exact result; negative = saldo a favor.
    @JsonKey(name: 'amount_cents') required MoneyCents amountCents,

    /// Amount to pay: the positive part of `amount_cents` (0 when `in_favor` or `zero`). Centavo precision until the CPA confirms CFF art. 20 whole-peso rounding (`rounding.payment_whole_pesos`, off, OQ-T3).
    @JsonKey(name: 'payable_rounded_cents')
    required NonNegativeCents payableRoundedCents,
    required TaxResultOutcome outcome,
  }) = _TaxResult;

  factory TaxResult.fromJson(Map<String, Object?> json) =>
      _$TaxResultFromJson(json);
}
