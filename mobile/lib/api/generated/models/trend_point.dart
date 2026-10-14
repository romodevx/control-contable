// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'period.dart';
import 'trend_tax.dart';

part 'trend_point.freezed.dart';
part 'trend_point.g.dart';

/// One month of the trend. Income and expense as in `DashboardMonth` (`null` ⇔ that module OFF); `tax` is `null` ⇔ `module_tax_engine` OFF. A month with no record is a point with zeros, never a hole.
@Freezed()
abstract class TrendPoint with _$TrendPoint {
  const factory TrendPoint({
    required Period period,
    @JsonKey(includeIfNull: true, name: 'income_cents')
    required int? incomeCents,
    @JsonKey(includeIfNull: true, name: 'expense_cents')
    required int? expenseCents,
    @JsonKey(includeIfNull: true) required TrendTax? tax,
  }) = _TrendPoint;

  factory TrendPoint.fromJson(Map<String, Object?> json) =>
      _$TrendPointFromJson(json);
}
