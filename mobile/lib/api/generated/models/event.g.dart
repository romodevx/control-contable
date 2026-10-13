// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'event.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Event _$EventFromJson(Map<String, dynamic> json) => _Event(
  id: json['id'] as String,
  type: json['type'] as String,
  appUserId: json['app_user_id'] as String,
  eventTimestampMs: (json['event_timestamp_ms'] as num).toInt(),
  originalAppUserId: json['original_app_user_id'] as String?,
  productId: json['product_id'] as String?,
  entitlementIds: (json['entitlement_ids'] as List<dynamic>?)
      ?.map((e) => e as String)
      .toList(),
  environment: json['environment'] == null
      ? null
      : Environment.fromJson(json['environment'] as String),
  store: json['store'] as String?,
  expirationAtMs: (json['expiration_at_ms'] as num?)?.toInt(),
);

Map<String, dynamic> _$EventToJson(_Event instance) => <String, dynamic>{
  'id': instance.id,
  'type': instance.type,
  'app_user_id': instance.appUserId,
  'event_timestamp_ms': instance.eventTimestampMs,
  'original_app_user_id': ?instance.originalAppUserId,
  'product_id': ?instance.productId,
  'entitlement_ids': ?instance.entitlementIds,
  'environment': ?_$EnvironmentEnumMap[instance.environment],
  'store': ?instance.store,
  'expiration_at_ms': ?instance.expirationAtMs,
};

const _$EnvironmentEnumMap = {
  Environment.sandbox: 'SANDBOX',
  Environment.production: 'PRODUCTION',
  Environment.$unknown: r'$unknown',
};
