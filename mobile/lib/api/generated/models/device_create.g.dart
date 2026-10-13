// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'device_create.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_DeviceCreate _$DeviceCreateFromJson(Map<String, dynamic> json) =>
    _DeviceCreate(
      id: json['id'] as String,
      platform: DeviceCreatePlatform.fromJson(json['platform'] as String),
      pushToken: json['push_token'] as String,
      appVersion: json['app_version'] as String?,
    );

Map<String, dynamic> _$DeviceCreateToJson(_DeviceCreate instance) =>
    <String, dynamic>{
      'id': instance.id,
      'platform': _$DeviceCreatePlatformEnumMap[instance.platform]!,
      'push_token': instance.pushToken,
      'app_version': instance.appVersion,
    };

const _$DeviceCreatePlatformEnumMap = {
  DeviceCreatePlatform.ios: 'ios',
  DeviceCreatePlatform.android: 'android',
  DeviceCreatePlatform.$unknown: r'$unknown',
};
