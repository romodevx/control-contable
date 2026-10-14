// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'access_level.dart';

part 'access_summary.freezed.dart';
part 'access_summary.g.dart';

/// The caller's access state, computed by the server so the client never uses its own clock (OQ-D6). Per user, across all taxpayers (PRD Q2). One shape for every screen that needs it: the Home banner (`getDashboard`), the subscription status (M7) and the check before a create wizard (MODULES.md M7, PR #58 follow-up: M7-01 exposes this same schema to the app regardless of `module_subscriptions`). `trial`: `trial_ends_at` (exclusive), `trial_days_left` (whole days left, rounded up, at least 1 while the trial runs) and the records left of each kind. `read_only`: both counters `0`, the trial fields `null`. `pro`, `comp`: every other field `null` (no limit applies).
@Freezed()
abstract class AccessSummary with _$AccessSummary {
  const factory AccessSummary({
    required AccessLevel access,
    @JsonKey(includeIfNull: true, name: 'trial_ends_at')
    required DateTime? trialEndsAt,
    @JsonKey(includeIfNull: true, name: 'trial_days_left')
    required int? trialDaysLeft,
    @JsonKey(includeIfNull: true, name: 'incomes_remaining')
    required int? incomesRemaining,
    @JsonKey(includeIfNull: true, name: 'expenses_remaining')
    required int? expensesRemaining,
  }) = _AccessSummary;

  factory AccessSummary.fromJson(Map<String, Object?> json) =>
      _$AccessSummaryFromJson(json);
}
