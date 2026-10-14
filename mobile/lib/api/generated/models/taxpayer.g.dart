// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'taxpayer.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Taxpayer _$TaxpayerFromJson(Map<String, dynamic> json) => _Taxpayer(
  id: json['id'] as String,
  personType: PersonType.fromJson(json['person_type'] as String),
  displayName: json['display_name'] as String,
  rfcMasked: json['rfc_masked'] as String,
  regimes: (json['regimes'] as List<dynamic>)
      .map((e) => TaxpayerRegime.fromJson(e as Map<String, dynamic>))
      .toList(),
  archivedAt: json['archived_at'] == null
      ? null
      : DateTime.parse(json['archived_at'] as String),
  createdAt: DateTime.parse(json['created_at'] as String),
  warnings: (json['warnings'] as List<dynamic>?)
      ?.map((e) => Warning.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$TaxpayerToJson(_Taxpayer instance) => <String, dynamic>{
  'id': instance.id,
  'person_type': _$PersonTypeEnumMap[instance.personType]!,
  'display_name': instance.displayName,
  'rfc_masked': instance.rfcMasked,
  'regimes': instance.regimes,
  'archived_at': instance.archivedAt?.toIso8601String(),
  'created_at': instance.createdAt.toIso8601String(),
  'warnings': ?instance.warnings,
};

const _$PersonTypeEnumMap = {
  PersonType.fisica: 'fisica',
  PersonType.moral: 'moral',
  PersonType.$unknown: r'$unknown',
};
