// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'access_summary.dart';
import 'subscription_plan.dart';

part 'subscription_status.freezed.dart';
part 'subscription_status.g.dart';

@Freezed()
abstract class SubscriptionStatus with _$SubscriptionStatus {
  const factory SubscriptionStatus({
    required AccessSummary access,
    @JsonKey(includeIfNull: false) SubscriptionPlan? plan,
  }) = _SubscriptionStatus;

  factory SubscriptionStatus.fromJson(Map<String, Object?> json) =>
      _$SubscriptionStatusFromJson(json);
}
