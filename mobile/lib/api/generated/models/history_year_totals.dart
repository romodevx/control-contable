// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'non_negative_cents.dart';

part 'history_year_totals.freezed.dart';
part 'history_year_totals.g.dart';

/// "Suma de estimados mensuales": the sum over the returned months of their `tax` figures (`null` months add nothing). A sum of provisional estimates, **not** the annual return; a saldo a favor never subtracts from it. `total_payable_rounded_cents = isr_payable_rounded_cents + iva_payable_rounded_cents + withholdings_to_remit_cents`.
@Freezed()
abstract class HistoryYearTotals with _$HistoryYearTotals {
  const factory HistoryYearTotals({
    @JsonKey(name: 'isr_payable_rounded_cents')
    required NonNegativeCents isrPayableRoundedCents,
    @JsonKey(name: 'iva_payable_rounded_cents')
    required NonNegativeCents ivaPayableRoundedCents,
    @JsonKey(name: 'withholdings_to_remit_cents')
    required NonNegativeCents withholdingsToRemitCents,
    @JsonKey(name: 'total_payable_rounded_cents')
    required NonNegativeCents totalPayableRoundedCents,
  }) = _HistoryYearTotals;

  factory HistoryYearTotals.fromJson(Map<String, Object?> json) =>
      _$HistoryYearTotalsFromJson(json);
}
