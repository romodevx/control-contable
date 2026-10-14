// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'regime_option_def.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_RegimeOptionDef _$RegimeOptionDefFromJson(Map<String, dynamic> json) =>
    _RegimeOptionDef(
      key: json['key'] as String,
      label: json['label'] as String,
      type: RegimeOptionDefType.fromJson(json['type'] as String),
      requiredValue: json['required'] as bool,
      allowedValues: (json['allowed_values'] as List<dynamic>?)
          ?.map((e) => e as String)
          .toList(),
      min: json['min'] as String?,
      max: json['max'] as String?,
      defaultValue: json['default'],
    );

Map<String, dynamic> _$RegimeOptionDefToJson(_RegimeOptionDef instance) =>
    <String, dynamic>{
      'key': instance.key,
      'label': instance.label,
      'type': _$RegimeOptionDefTypeEnumMap[instance.type]!,
      'required': instance.requiredValue,
      'allowed_values': ?instance.allowedValues,
      'min': ?instance.min,
      'max': ?instance.max,
      'default': ?instance.defaultValue,
    };

const _$RegimeOptionDefTypeEnumMap = {
  RegimeOptionDefType.valueEnum: 'enum',
  RegimeOptionDefType.multiEnum: 'multi_enum',
  RegimeOptionDefType.boolean: 'boolean',
  RegimeOptionDefType.rate: 'rate',
  RegimeOptionDefType.$unknown: r'$unknown',
};
