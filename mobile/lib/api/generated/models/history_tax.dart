// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'history_tax_iva_outcome.dart';
import 'non_negative_cents.dart';

part 'history_tax.freezed.dart';
part 'history_tax.g.dart';

/// One month's taxpayer-level figures, from the same estimate `getTaxpayerEstimate` answers for that month (`totals`), to the centavo. A saldo a favor is shown in words (`iva_outcome: in_favor`, `iva_in_favor_cents` > 0) and never adds to a total: `iva_payable_rounded_cents` is then `0`. `total_payable_rounded_cents = isr_payable_rounded_cents + iva_payable_rounded_cents + withholdings_to_remit_cents`.
@Freezed()
abstract class HistoryTax with _$HistoryTax {
  const factory HistoryTax({
    @JsonKey(name: 'isr_payable_rounded_cents')
    required NonNegativeCents isrPayableRoundedCents,
    @JsonKey(name: 'iva_payable_rounded_cents')
    required NonNegativeCents ivaPayableRoundedCents,
    @JsonKey(name: 'iva_outcome') required HistoryTaxIvaOutcome ivaOutcome,

    /// Saldo a favor of the consolidated IVA (the magnitude of a negative `totals.iva.amount_cents`); `0` unless `iva_outcome` is `in_favor`.
    @JsonKey(name: 'iva_in_favor_cents')
    required NonNegativeCents ivaInFavorCents,
    @JsonKey(name: 'withholdings_to_remit_cents')
    required NonNegativeCents withholdingsToRemitCents,
    @JsonKey(name: 'total_payable_rounded_cents')
    required NonNegativeCents totalPayableRoundedCents,
  }) = _HistoryTax;

  factory HistoryTax.fromJson(Map<String, Object?> json) =>
      _$HistoryTaxFromJson(json);
}
