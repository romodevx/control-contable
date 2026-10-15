// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'me.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Me _$MeFromJson(Map<String, dynamic> json) => _Me(
  id: json['id'] as String,
  email: json['email'] as String,
  createdAt: DateTime.parse(json['created_at'] as String),
  access: AccessSummary.fromJson(json['access'] as Map<String, dynamic>),
  name: json['name'] as String?,
);

Map<String, dynamic> _$MeToJson(_Me instance) => <String, dynamic>{
  'id': instance.id,
  'email': instance.email,
  'created_at': instance.createdAt.toIso8601String(),
  'access': instance.access,
  'name': ?instance.name,
};
