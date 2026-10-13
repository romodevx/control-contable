// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'verify_email_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_VerifyEmailRequest _$VerifyEmailRequestFromJson(Map<String, dynamic> json) =>
    _VerifyEmailRequest(
      email: json['email'] as String,
      code: json['code'] as String,
    );

Map<String, dynamic> _$VerifyEmailRequestToJson(_VerifyEmailRequest instance) =>
    <String, dynamic>{'email': instance.email, 'code': instance.code};
