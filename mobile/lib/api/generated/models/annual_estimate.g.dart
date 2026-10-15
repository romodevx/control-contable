// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'annual_estimate.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_AnnualEstimate _$AnnualEstimateFromJson(Map<String, dynamic> json) =>
    _AnnualEstimate(
      taxpayerId: json['taxpayer_id'] as String,
      fiscalYear: (json['fiscal_year'] as num).toInt(),
      personType: PersonType.fromJson(json['person_type'] as String),
      legend: json['legend'] as String,
      disclaimers: (json['disclaimers'] as List<dynamic>)
          .map((e) => Disclaimer.fromJson(e as Map<String, dynamic>))
          .toList(),
      yearStatus: AnnualYearStatus.fromJson(json['year_status'] as String),
      dataThrough: json['data_through'] as String?,
      deadline: json['deadline'] as String?,
      headlineSections: (json['headline_sections'] as List<dynamic>)
          .map((e) => AnnualComputedSectionKind.fromJson(e as String))
          .toList(),
      sections: (json['sections'] as List<dynamic>)
          .map((e) => AnnualSection.fromJson(e as Map<String, dynamic>))
          .toList(),
      informativeIncomeOver500k: json['informative_income_over_500k'] as bool,
      parameterSet: ParameterSetRef.fromJson(
        json['parameter_set'] as Map<String, dynamic>,
      ),
      engineVersion: json['engine_version'] as String,
    );

Map<String, dynamic> _$AnnualEstimateToJson(_AnnualEstimate instance) =>
    <String, dynamic>{
      'taxpayer_id': instance.taxpayerId,
      'fiscal_year': instance.fiscalYear,
      'person_type': _$PersonTypeEnumMap[instance.personType]!,
      'legend': instance.legend,
      'disclaimers': instance.disclaimers,
      'year_status': _$AnnualYearStatusEnumMap[instance.yearStatus]!,
      'data_through': instance.dataThrough,
      'deadline': instance.deadline,
      'headline_sections': instance.headlineSections
          .map((e) => _$AnnualComputedSectionKindEnumMap[e]!)
          .toList(),
      'sections': instance.sections,
      'informative_income_over_500k': instance.informativeIncomeOver500k,
      'parameter_set': instance.parameterSet,
      'engine_version': instance.engineVersion,
    };

const _$PersonTypeEnumMap = {
  PersonType.fisica: 'fisica',
  PersonType.moral: 'moral',
  PersonType.$unknown: r'$unknown',
};

const _$AnnualYearStatusEnumMap = {
  AnnualYearStatus.closed: 'closed',
  AnnualYearStatus.inProgress: 'in_progress',
  AnnualYearStatus.$unknown: r'$unknown',
};

const _$AnnualComputedSectionKindEnumMap = {
  AnnualComputedSectionKind.pfArt152: 'pf_art152',
  AnnualComputedSectionKind.pfResicoCheck: 'pf_resico_check',
  AnnualComputedSectionKind.pm601: 'pm_601',
  AnnualComputedSectionKind.pmResico: 'pm_resico',
  AnnualComputedSectionKind.$unknown: r'$unknown',
};
