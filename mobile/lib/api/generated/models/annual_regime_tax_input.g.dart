// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'annual_regime_tax_input.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_AnnualRegimeTaxInput _$AnnualRegimeTaxInputFromJson(
  Map<String, dynamic> json,
) => _AnnualRegimeTaxInput(
  regimeCode: json['regime_code'] as String,
  key: json['key'] as String,
  label: json['label'] as String,
  valueCents: (json['value_cents'] as num?)?.toInt(),
);

Map<String, dynamic> _$AnnualRegimeTaxInputToJson(
  _AnnualRegimeTaxInput instance,
) => <String, dynamic>{
  'regime_code': instance.regimeCode,
  'key': instance.key,
  'label': instance.label,
  'value_cents': instance.valueCents,
};
