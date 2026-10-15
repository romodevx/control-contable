// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'annual_inputs_update.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_AnnualInputsUpdate _$AnnualInputsUpdateFromJson(Map<String, dynamic> json) =>
    _AnnualInputsUpdate(
      priorYearAccumulableIncomeCents:
          (json['prior_year_accumulable_income_cents'] as num?)?.toInt(),
      otherExemptIncomeCents: (json['other_exempt_income_cents'] as num)
          .toInt(),
      toldEmployerWillFile: json['told_employer_will_file'] as bool,
    );

Map<String, dynamic> _$AnnualInputsUpdateToJson(_AnnualInputsUpdate instance) =>
    <String, dynamic>{
      'prior_year_accumulable_income_cents':
          instance.priorYearAccumulableIncomeCents,
      'other_exempt_income_cents': instance.otherExemptIncomeCents,
      'told_employer_will_file': instance.toldEmployerWillFile,
    };
