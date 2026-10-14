// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'tax_input_value.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_TaxInputValue _$TaxInputValueFromJson(Map<String, dynamic> json) =>
    _TaxInputValue(
      key: json['key'] as String,
      month: (json['month'] as num?)?.toInt(),
      valueCents: (json['value_cents'] as num?)?.toInt(),
      valueRate: json['value_rate'] as String?,
      valueBool: json['value_bool'] as bool?,
    );

Map<String, dynamic> _$TaxInputValueToJson(_TaxInputValue instance) =>
    <String, dynamic>{
      'key': instance.key,
      'month': ?instance.month,
      'value_cents': ?instance.valueCents,
      'value_rate': ?instance.valueRate,
      'value_bool': ?instance.valueBool,
    };
