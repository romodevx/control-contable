// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'token_pair.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_TokenPair _$TokenPairFromJson(Map<String, dynamic> json) => _TokenPair(
  accessToken: json['access_token'] as String,
  accessExpiresIn: (json['access_expires_in'] as num).toInt(),
  refreshToken: json['refresh_token'] as String,
  refreshExpiresIn: (json['refresh_expires_in'] as num).toInt(),
  tokenType: json['token_type'] as String,
);

Map<String, dynamic> _$TokenPairToJson(_TokenPair instance) =>
    <String, dynamic>{
      'access_token': instance.accessToken,
      'access_expires_in': instance.accessExpiresIn,
      'refresh_token': instance.refreshToken,
      'refresh_expires_in': instance.refreshExpiresIn,
      'token_type': instance.tokenType,
    };
