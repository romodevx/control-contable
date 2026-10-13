// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'non_negative_cents.dart';

part 'period_totals.freezed.dart';
part 'period_totals.g.dart';

@Freezed()
abstract class PeriodTotals with _$PeriodTotals {
  const factory PeriodTotals({
    required int count,
    @JsonKey(name: 'amount_cents') required NonNegativeCents amountCents,
    @JsonKey(name: 'iva_cents') required NonNegativeCents ivaCents,
    @JsonKey(name: 'isr_withheld_cents') NonNegativeCents? isrWithheldCents,
    @JsonKey(name: 'iva_withheld_cents') NonNegativeCents? ivaWithheldCents,
  }) = _PeriodTotals;

  factory PeriodTotals.fromJson(Map<String, Object?> json) =>
      _$PeriodTotalsFromJson(json);
}
