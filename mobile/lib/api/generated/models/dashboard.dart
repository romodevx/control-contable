// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'access_summary.dart';
import 'legend.dart';
import 'non_negative_cents.dart';
import 'period.dart';
import 'taxpayer_estimate.dart';
import 'trend_point.dart';
import 'uuid.dart';

part 'dashboard.freezed.dart';
part 'dashboard.g.dart';

@Freezed()
abstract class Dashboard with _$Dashboard {
  const factory Dashboard({
    @JsonKey(name: 'taxpayer_id') required Uuid taxpayerId,
    required Period period,

    /// True when the taxpayer has no records (empty state).
    @JsonKey(name: 'is_empty') required bool isEmpty,
    required TaxpayerEstimate estimate,
    @JsonKey(name: 'income_cents') required NonNegativeCents incomeCents,
    @JsonKey(name: 'expense_cents') required NonNegativeCents expenseCents,
    required List<TrendPoint> trend,
    required AccessSummary access,
    required Legend legend,
  }) = _Dashboard;

  factory Dashboard.fromJson(Map<String, Object?> json) =>
      _$DashboardFromJson(json);
}
