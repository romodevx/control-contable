// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'enabled_flags.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_EnabledFlags _$EnabledFlagsFromJson(Map<String, dynamic> json) =>
    _EnabledFlags(
      enabled: (json['enabled'] as List<dynamic>)
          .map((e) => e as String)
          .toList(),
    );

Map<String, dynamic> _$EnabledFlagsToJson(_EnabledFlags instance) =>
    <String, dynamic>{'enabled': instance.enabled};
