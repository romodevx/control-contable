// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'history_month_status.dart';
import 'non_negative_cents.dart';
import 'period.dart';
import 'tax_result.dart';

part 'history_month.freezed.dart';
part 'history_month.g.dart';

@Freezed()
abstract class HistoryMonth with _$HistoryMonth {
  const factory HistoryMonth({
    required Period period,
    required HistoryMonthStatus status,
    @JsonKey(name: 'total_payable_rounded_cents')
    required NonNegativeCents totalPayableRoundedCents,
    @JsonKey(name: 'changed_since_viewed') required bool changedSinceViewed,
    @JsonKey(includeIfNull: false) TaxResult? iva,
    @JsonKey(includeIfNull: false) TaxResult? isr,
    @JsonKey(includeIfNull: false, name: 'last_viewed_at')
    DateTime? lastViewedAt,
  }) = _HistoryMonth;

  factory HistoryMonth.fromJson(Map<String, Object?> json) =>
      _$HistoryMonthFromJson(json);
}
