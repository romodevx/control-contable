// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'missing_input.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_MissingInput _$MissingInputFromJson(Map<String, dynamic> json) =>
    _MissingInput(
      key: json['key'] as String,
      label: json['label'] as String,
      months: (json['months'] as List<dynamic>?)
          ?.map((e) => e as String)
          .toList(),
    );

Map<String, dynamic> _$MissingInputToJson(_MissingInput instance) =>
    <String, dynamic>{
      'key': instance.key,
      'label': instance.label,
      'months': ?instance.months,
    };
