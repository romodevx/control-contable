// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'annual_missing_input.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_AnnualMissingInput _$AnnualMissingInputFromJson(Map<String, dynamic> json) =>
    _AnnualMissingInput(
      key: json['key'] as String,
      label: json['label'] as String,
      target: AnnualInputTarget.fromJson(json['target'] as String),
      regimeCode: json['regime_code'] as String?,
      months: (json['months'] as List<dynamic>)
          .map((e) => e as String)
          .toList(),
    );

Map<String, dynamic> _$AnnualMissingInputToJson(_AnnualMissingInput instance) =>
    <String, dynamic>{
      'key': instance.key,
      'label': instance.label,
      'target': _$AnnualInputTargetEnumMap[instance.target]!,
      'regime_code': instance.regimeCode,
      'months': instance.months,
    };

const _$AnnualInputTargetEnumMap = {
  AnnualInputTarget.salaryStatements: 'salary_statements',
  AnnualInputTarget.personalDeductions: 'personal_deductions',
  AnnualInputTarget.annualInputs: 'annual_inputs',
  AnnualInputTarget.filedPayments: 'filed_payments',
  AnnualInputTarget.taxInputs: 'tax_inputs',
  AnnualInputTarget.$unknown: r'$unknown',
};
