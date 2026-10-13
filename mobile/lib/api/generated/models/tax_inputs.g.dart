// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'tax_inputs.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_TaxInputs _$TaxInputsFromJson(Map<String, dynamic> json) => _TaxInputs(
  regimeCode: json['regime_code'] as String,
  fiscalYear: (json['fiscal_year'] as num).toInt(),
  definitions: (json['definitions'] as List<dynamic>)
      .map((e) => TaxInputDef.fromJson(e as Map<String, dynamic>))
      .toList(),
  values: (json['values'] as List<dynamic>)
      .map((e) => TaxInputValue.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$TaxInputsToJson(_TaxInputs instance) =>
    <String, dynamic>{
      'regime_code': instance.regimeCode,
      'fiscal_year': instance.fiscalYear,
      'definitions': instance.definitions,
      'values': instance.values,
    };
