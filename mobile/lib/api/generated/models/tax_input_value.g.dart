// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'tax_input_value.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_TaxInputValue _$TaxInputValueFromJson(Map<String, dynamic> json) =>
    _TaxInputValue(
      key: json['key'] as String,
      valueCents: (json['value_cents'] as num?)?.toInt(),
      valueRate: json['value_rate'] as String?,
    );

Map<String, dynamic> _$TaxInputValueToJson(_TaxInputValue instance) =>
    <String, dynamic>{
      'key': instance.key,
      'value_cents': instance.valueCents,
      'value_rate': instance.valueRate,
    };
