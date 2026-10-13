// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'subscription_status.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_SubscriptionStatus _$SubscriptionStatusFromJson(Map<String, dynamic> json) =>
    _SubscriptionStatus(
      access: AccessSummary.fromJson(json['access'] as Map<String, dynamic>),
      plan: json['plan'] == null
          ? null
          : SubscriptionPlan.fromJson(json['plan'] as Map<String, dynamic>),
    );

Map<String, dynamic> _$SubscriptionStatusToJson(_SubscriptionStatus instance) =>
    <String, dynamic>{'access': instance.access, 'plan': ?instance.plan};
