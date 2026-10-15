// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'annual_inputs.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_AnnualInputs _$AnnualInputsFromJson(Map<String, dynamic> json) =>
    _AnnualInputs(
      fiscalYear: (json['fiscal_year'] as num).toInt(),
      priorYearAccumulableIncomeCents:
          (json['prior_year_accumulable_income_cents'] as num?)?.toInt(),
      otherExemptIncomeCents: (json['other_exempt_income_cents'] as num)
          .toInt(),
      toldEmployerWillFile: json['told_employer_will_file'] as bool,
      regimeTaxInputs: (json['regime_tax_inputs'] as List<dynamic>)
          .map((e) => AnnualRegimeTaxInput.fromJson(e as Map<String, dynamic>))
          .toList(),
      updatedAt: json['updated_at'] == null
          ? null
          : DateTime.parse(json['updated_at'] as String),
    );

Map<String, dynamic> _$AnnualInputsToJson(_AnnualInputs instance) =>
    <String, dynamic>{
      'fiscal_year': instance.fiscalYear,
      'prior_year_accumulable_income_cents':
          instance.priorYearAccumulableIncomeCents,
      'other_exempt_income_cents': instance.otherExemptIncomeCents,
      'told_employer_will_file': instance.toldEmployerWillFile,
      'regime_tax_inputs': instance.regimeTaxInputs,
      'updated_at': instance.updatedAt?.toIso8601String(),
    };
