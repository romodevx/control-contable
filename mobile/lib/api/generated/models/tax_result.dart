// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'money_cents.dart';
import 'tax_result_outcome.dart';

part 'tax_result.freezed.dart';
part 'tax_result.g.dart';

@Freezed()
abstract class TaxResult with _$TaxResult {
  const factory TaxResult({
    /// Exact result; negative = saldo a favor.
    @JsonKey(name: 'amount_cents') required MoneyCents amountCents,

    /// Payment amount. Equals amount_cents (centavo precision) until the CPA confirms CFF art. 20 whole-peso rounding (ADR-0002 §4).
    @JsonKey(name: 'payable_rounded_cents')
    required MoneyCents payableRoundedCents,
    required TaxResultOutcome outcome,
  }) = _TaxResult;

  factory TaxResult.fromJson(Map<String, Object?> json) =>
      _$TaxResultFromJson(json);
}
