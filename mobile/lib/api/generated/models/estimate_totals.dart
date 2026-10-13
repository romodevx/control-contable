// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'non_negative_cents.dart';
import 'tax_result.dart';

part 'estimate_totals.freezed.dart';
part 'estimate_totals.g.dart';

@Freezed()
abstract class EstimateTotals with _$EstimateTotals {
  const factory EstimateTotals({
    /// Monthly IVA consolidated across the taxpayer's regimes.
    required TaxResult iva,

    /// Sum of per-regime provisional ISR (payable parts only; favor balances per regime).
    required TaxResult isr,

    /// Withheld ISR/IVA to remit in full (not netted against IVA, LIVA 1-A).
    @JsonKey(name: 'withholdings_to_remit_cents')
    required NonNegativeCents withholdingsToRemitCents,
    @JsonKey(name: 'total_payable_rounded_cents')
    required NonNegativeCents totalPayableRoundedCents,
  }) = _EstimateTotals;

  factory EstimateTotals.fromJson(Map<String, Object?> json) =>
      _$EstimateTotalsFromJson(json);
}
