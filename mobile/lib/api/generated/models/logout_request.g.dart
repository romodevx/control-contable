// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'logout_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_LogoutRequest _$LogoutRequestFromJson(Map<String, dynamic> json) =>
    _LogoutRequest(
      refreshToken: json['refresh_token'] as String,
      deviceId: json['device_id'] as String?,
    );

Map<String, dynamic> _$LogoutRequestToJson(_LogoutRequest instance) =>
    <String, dynamic>{
      'refresh_token': instance.refreshToken,
      'device_id': ?instance.deviceId,
    };
