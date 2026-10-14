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
  explanation: json['explanation'] as String?,
  optionDefs: (json['option_defs'] as List<dynamic>)
      .map((e) => RegimeOptionDef.fromJson(e as Map<String, dynamic>))
      .toList(),
  requiredInputs: (json['required_inputs'] as List<dynamic>)
      .map((e) => TaxInputDef.fromJson(e as Map<String, dynamic>))
      .toList(),
  incompatibleWith: (json['incompatible_with'] as List<dynamic>)
      .map((e) => RegimeIncompatibility.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$RegimeToJson(_Regime instance) => <String, dynamic>{
  'code': instance.code,
  'name': instance.name,
  'person_type': _$PersonTypeEnumMap[instance.personType]!,
  'treatment': _$TreatmentEnumMap[instance.treatment]!,
  'explanation': instance.explanation,
  'option_defs': instance.optionDefs,
  'required_inputs': instance.requiredInputs,
  'incompatible_with': instance.incompatibleWith,
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
