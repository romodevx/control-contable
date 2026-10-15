// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'device.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Device _$DeviceFromJson(Map<String, dynamic> json) => _Device(
  id: json['id'] as String,
  platform: DevicePlatform.fromJson(json['platform'] as String),
  appVersion: json['app_version'] as String?,
  createdAt: DateTime.parse(json['created_at'] as String),
  lastSeenAt: DateTime.parse(json['last_seen_at'] as String),
);

Map<String, dynamic> _$DeviceToJson(_Device instance) => <String, dynamic>{
  'id': instance.id,
  'platform': _$DevicePlatformEnumMap[instance.platform]!,
  'app_version': instance.appVersion,
  'created_at': instance.createdAt.toIso8601String(),
  'last_seen_at': instance.lastSeenAt.toIso8601String(),
};

const _$DevicePlatformEnumMap = {
  DevicePlatform.ios: 'ios',
  DevicePlatform.android: 'android',
  DevicePlatform.$unknown: r'$unknown',
};
