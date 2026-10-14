// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'tax_input_def.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_TaxInputDef _$TaxInputDefFromJson(Map<String, dynamic> json) => _TaxInputDef(
  key: json['key'] as String,
  label: json['label'] as String,
  kind: TaxInputDefKind.fromJson(json['kind'] as String),
  requiredValue: json['required'] as bool,
);

Map<String, dynamic> _$TaxInputDefToJson(_TaxInputDef instance) =>
    <String, dynamic>{
      'key': instance.key,
      'label': instance.label,
      'kind': _$TaxInputDefKindEnumMap[instance.kind]!,
      'required': instance.requiredValue,
    };

const _$TaxInputDefKindEnumMap = {
  TaxInputDefKind.money: 'money',
  TaxInputDefKind.moneyMonthly: 'money_monthly',
  TaxInputDefKind.rate: 'rate',
  TaxInputDefKind.boolean: 'boolean',
  TaxInputDefKind.$unknown: r'$unknown',
};
