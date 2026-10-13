// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'regime.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Regime _$RegimeFromJson(Map<String, dynamic> json) => _Regime(
  code: json['code'] as String,
  name: json['name'] as String,
  personType: PersonType.fromJson(json['person_type'] as String),
  treatment: Treatment.fromJson(json['treatment'] as String),
  optionDefs: (json['option_defs'] as List<dynamic>)
      .map((e) => RegimeOptionDef.fromJson(e as Map<String, dynamic>))
      .toList(),
  requiredInputs: (json['required_inputs'] as List<dynamic>)
      .map((e) => e as String)
      .toList(),
  explanation: json['explanation'] as String?,
);

Map<String, dynamic> _$RegimeToJson(_Regime instance) => <String, dynamic>{
  'code': instance.code,
  'name': instance.name,
  'person_type': _$PersonTypeEnumMap[instance.personType]!,
  'treatment': _$TreatmentEnumMap[instance.treatment]!,
  'option_defs': instance.optionDefs,
  'required_inputs': instance.requiredInputs,
  'explanation': ?instance.explanation,
};

const _$PersonTypeEnumMap = {
  PersonType.fisica: 'fisica',
  PersonType.moral: 'moral',
  PersonType.$unknown: r'$unknown',
};

const _$TreatmentEnumMap = {
  Treatment.full: 'full',
  Treatment.partial: 'partial',
  Treatment.informational: 'informational',
  Treatment.$unknown: r'$unknown',
};
