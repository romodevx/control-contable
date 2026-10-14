// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'access_summary.dart';
import 'dashboard_month.dart';
import 'legend.dart';
import 'period.dart';
import 'taxpayer_estimate.dart';
import 'trend_point.dart';
import 'uuid.dart';

part 'dashboard.freezed.dart';
part 'dashboard.g.dart';

/// Home of one taxpayer for `period` (see `getDashboard`). `is_empty` is true when the taxpayer has no live income and no live expense in any period (enabled modules only): the client shows the empty state instead of the hero and the chart. `estimate` is `null` ⇔ `module_tax_engine` OFF. `trend` is oldest first and ends at `period`. `legend` is the mandatory constant, present if and only if `estimate` is not `null`; the client renders it under every tax figure, never a copy of its own. The `strong` disclaimers (`cpa_review_pending`) travel inside `estimate.disclaimers`.
@Freezed()
abstract class Dashboard with _$Dashboard {
  const factory Dashboard({
    @JsonKey(name: 'taxpayer_id') required Uuid taxpayerId,
    required Period period,
    @JsonKey(name: 'is_empty') required bool isEmpty,
    required DashboardMonth month,
    @JsonKey(includeIfNull: true) required TaxpayerEstimate? estimate,
    required List<TrendPoint> trend,
    required AccessSummary access,
    @JsonKey(includeIfNull: false) Legend? legend,
  }) = _Dashboard;

  factory Dashboard.fromJson(Map<String, Object?> json) =>
      _$DashboardFromJson(json);
}
