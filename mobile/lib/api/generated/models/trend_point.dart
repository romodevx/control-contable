// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'non_negative_cents.dart';
import 'period.dart';

part 'trend_point.freezed.dart';
part 'trend_point.g.dart';

@Freezed()
abstract class TrendPoint with _$TrendPoint {
  const factory TrendPoint({
    required Period period,
    @JsonKey(name: 'income_cents') required NonNegativeCents incomeCents,
    @JsonKey(name: 'expense_cents') required NonNegativeCents expenseCents,
    @JsonKey(name: 'tax_payable_rounded_cents')
    required NonNegativeCents taxPayableRoundedCents,
  }) = _TrendPoint;

  factory TrendPoint.fromJson(Map<String, Object?> json) =>
      _$TrendPointFromJson(json);
}
