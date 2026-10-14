// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'non_negative_cents.dart';
import 'taxpayer_estimate_status.dart';
import 'trend_tax_iva_outcome.dart';

part 'trend_tax.freezed.dart';
part 'trend_tax.g.dart';

/// One month's tax figures, from the same estimate `getTaxpayerEstimate` answers for that month (`totals`). A saldo a favor never adds to a bar: `iva_payable_rounded_cents` is `0` and `iva_outcome` is `in_favor`. `total_payable_rounded_cents = iva_payable_rounded_cents + isr_payable_rounded_cents + withholdings_to_remit_cents`.
@Freezed()
abstract class TrendTax with _$TrendTax {
  const factory TrendTax({
    @JsonKey(name: 'iva_payable_rounded_cents')
    required NonNegativeCents ivaPayableRoundedCents,
    @JsonKey(name: 'isr_payable_rounded_cents')
    required NonNegativeCents isrPayableRoundedCents,
    @JsonKey(name: 'withholdings_to_remit_cents')
    required NonNegativeCents withholdingsToRemitCents,
    @JsonKey(name: 'total_payable_rounded_cents')
    required NonNegativeCents totalPayableRoundedCents,
    @JsonKey(name: 'iva_outcome') required TrendTaxIvaOutcome ivaOutcome,
    required TaxpayerEstimateStatus status,
  }) = _TrendTax;

  factory TrendTax.fromJson(Map<String, Object?> json) =>
      _$TrendTaxFromJson(json);
}
