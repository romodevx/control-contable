// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'non_negative_cents.dart';
import 'tax_result.dart';

part 'estimate_totals.freezed.dart';
part 'estimate_totals.g.dart';

/// Taxpayer-level totals; they carry no status (the client shows "Incompleto" when any regime is `incomplete`). `total_payable_rounded_cents = isr.payable_rounded_cents + iva.payable_rounded_cents + withholdings_to_remit_cents`.
@Freezed()
abstract class EstimateTotals with _$EstimateTotals {
  const factory EstimateTotals({
    /// Monthly IVA consolidated across the taxpayer's regimes.
    required TaxResult iva,

    /// Sum of the payable per-regime provisional ISR; favor balances stay per regime (OQ-T10).
    required TaxResult isr,

    /// Withheld ISR/IVA to remit in full, shown on its own line (never netted against IVA, LIVA 1-A).
    @JsonKey(name: 'withholdings_to_remit_cents')
    required NonNegativeCents withholdingsToRemitCents,
    @JsonKey(name: 'total_payable_rounded_cents')
    required NonNegativeCents totalPayableRoundedCents,
  }) = _EstimateTotals;

  factory EstimateTotals.fromJson(Map<String, Object?> json) =>
      _$EstimateTotalsFromJson(json);
}
