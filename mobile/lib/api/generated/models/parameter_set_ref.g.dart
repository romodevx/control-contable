// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'parameter_set_ref.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_ParameterSetRef _$ParameterSetRefFromJson(Map<String, dynamic> json) =>
    _ParameterSetRef(
      fiscalYear: (json['fiscal_year'] as num).toInt(),
      version: (json['version'] as num).toInt(),
    );

Map<String, dynamic> _$ParameterSetRefToJson(_ParameterSetRef instance) =>
    <String, dynamic>{
      'fiscal_year': instance.fiscalYear,
      'version': instance.version,
    };
