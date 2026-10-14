// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

part 'dashboard_month.freezed.dart';
part 'dashboard_month.g.dart';

/// Income collected (`collected_on`) and expenses paid (`paid_on`) in the dashboard's `period`, before IVA, live records only. A field is `null` when its module (`module_incomes` / `module_expenses`) is effectively OFF; `0` when ON with no record.
@Freezed()
abstract class DashboardMonth with _$DashboardMonth {
  const factory DashboardMonth({
    @JsonKey(includeIfNull: true, name: 'income_cents')
    required int? incomeCents,
    @JsonKey(includeIfNull: true, name: 'expense_cents')
    required int? expenseCents,
  }) = _DashboardMonth;

  factory DashboardMonth.fromJson(Map<String, Object?> json) =>
      _$DashboardMonthFromJson(json);
}
