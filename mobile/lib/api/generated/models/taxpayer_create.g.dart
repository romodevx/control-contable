// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'taxpayer_create.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_TaxpayerCreate _$TaxpayerCreateFromJson(Map<String, dynamic> json) =>
    _TaxpayerCreate(
      personType: PersonType.fromJson(json['person_type'] as String),
      rfc: json['rfc'] as String,
      displayName: json['display_name'] as String,
      regimes: (json['regimes'] as List<dynamic>)
          .map((e) => TaxpayerRegimeCreate.fromJson(e as Map<String, dynamic>))
          .toList(),
      id: json['id'] as String?,
    );

Map<String, dynamic> _$TaxpayerCreateToJson(_TaxpayerCreate instance) =>
    <String, dynamic>{
      'person_type': _$PersonTypeEnumMap[instance.personType]!,
      'rfc': instance.rfc,
      'display_name': instance.displayName,
      'regimes': instance.regimes,
      'id': instance.id,
    };

const _$PersonTypeEnumMap = {
  PersonType.fisica: 'fisica',
  PersonType.moral: 'moral',
  PersonType.$unknown: r'$unknown',
};
