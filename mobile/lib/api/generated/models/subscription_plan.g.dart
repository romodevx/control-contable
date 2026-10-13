// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'subscription_plan.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_SubscriptionPlan _$SubscriptionPlanFromJson(Map<String, dynamic> json) =>
    _SubscriptionPlan(
      productId: json['product_id'] as String,
      period: SubscriptionPlanPeriod.fromJson(json['period'] as String),
      store: SubscriptionPlanStore.fromJson(json['store'] as String),
      expiresAt: DateTime.parse(json['expires_at'] as String),
      willRenew: json['will_renew'] as bool,
      managementUrl: json['management_url'] as String?,
    );

Map<String, dynamic> _$SubscriptionPlanToJson(_SubscriptionPlan instance) =>
    <String, dynamic>{
      'product_id': instance.productId,
      'period': _$SubscriptionPlanPeriodEnumMap[instance.period]!,
      'store': _$SubscriptionPlanStoreEnumMap[instance.store]!,
      'expires_at': instance.expiresAt.toIso8601String(),
      'will_renew': instance.willRenew,
      'management_url': instance.managementUrl,
    };

const _$SubscriptionPlanPeriodEnumMap = {
  SubscriptionPlanPeriod.weekly: 'weekly',
  SubscriptionPlanPeriod.monthly: 'monthly',
  SubscriptionPlanPeriod.yearly: 'yearly',
  SubscriptionPlanPeriod.$unknown: r'$unknown',
};

const _$SubscriptionPlanStoreEnumMap = {
  SubscriptionPlanStore.appStore: 'app_store',
  SubscriptionPlanStore.playStore: 'play_store',
  SubscriptionPlanStore.$unknown: r'$unknown',
};
