// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'access_summary_access.dart';

part 'access_summary.freezed.dart';
part 'access_summary.g.dart';

@Freezed()
abstract class AccessSummary with _$AccessSummary {
  const factory AccessSummary({
    required AccessSummaryAccess access,
    @JsonKey(name: 'trial_ends_at') DateTime? trialEndsAt,
    @JsonKey(name: 'incomes_remaining') int? incomesRemaining,
    @JsonKey(name: 'expenses_remaining') int? expensesRemaining,
  }) = _AccessSummary;

  factory AccessSummary.fromJson(Map<String, Object?> json) =>
      _$AccessSummaryFromJson(json);
}
