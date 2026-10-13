// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'revenue_cat_webhook.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_RevenueCatWebhook _$RevenueCatWebhookFromJson(Map<String, dynamic> json) =>
    _RevenueCatWebhook(
      apiVersion: json['api_version'] as String,
      event: Event.fromJson(json['event'] as Map<String, dynamic>),
    );

Map<String, dynamic> _$RevenueCatWebhookToJson(_RevenueCatWebhook instance) =>
    <String, dynamic>{
      'api_version': instance.apiVersion,
      'event': instance.event,
    };
