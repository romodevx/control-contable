// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'device_create.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_DeviceCreate _$DeviceCreateFromJson(Map<String, dynamic> json) =>
    _DeviceCreate(
      id: json['id'] as String,
      platform: DevicePlatform.fromJson(json['platform'] as String),
      pushToken: json['push_token'] as String,
      appVersion: json['app_version'] as String?,
    );

Map<String, dynamic> _$DeviceCreateToJson(_DeviceCreate instance) =>
    <String, dynamic>{
      'id': instance.id,
      'platform': _$DevicePlatformEnumMap[instance.platform]!,
      'push_token': instance.pushToken,
      'app_version': ?instance.appVersion,
    };

const _$DevicePlatformEnumMap = {
  DevicePlatform.ios: 'ios',
  DevicePlatform.android: 'android',
  DevicePlatform.$unknown: r'$unknown',
};
