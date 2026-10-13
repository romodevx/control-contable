// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'subscription_plan_period.dart';
import 'subscription_plan_store.dart';

part 'subscription_plan.freezed.dart';
part 'subscription_plan.g.dart';

/// Active store plan. A named component (not inline in `SubscriptionStatus.plan`):.
/// inline objects inside `oneOf` make the Dart generator name their enums after the.
/// bare property, and `period` overwrote the `Period` component (ADR-0003).
///
@Freezed()
abstract class SubscriptionPlan with _$SubscriptionPlan {
  const factory SubscriptionPlan({
    @JsonKey(name: 'product_id') required String productId,
    required SubscriptionPlanPeriod period,
    required SubscriptionPlanStore store,
    @JsonKey(name: 'expires_at') required DateTime expiresAt,
    @JsonKey(name: 'will_renew') required bool willRenew,
    @JsonKey(includeIfNull: false, name: 'management_url')
    String? managementUrl,
  }) = _SubscriptionPlan;

  factory SubscriptionPlan.fromJson(Map<String, Object?> json) =>
      _$SubscriptionPlanFromJson(json);
}
