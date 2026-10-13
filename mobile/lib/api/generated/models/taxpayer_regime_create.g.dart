// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'taxpayer_regime_create.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_TaxpayerRegimeCreate _$TaxpayerRegimeCreateFromJson(
  Map<String, dynamic> json,
) => _TaxpayerRegimeCreate(
  regimeCode: json['regime_code'] as String,
  options: json['options'] as Map<String, dynamic>?,
  startPeriod: json['start_period'] as String?,
);

Map<String, dynamic> _$TaxpayerRegimeCreateToJson(
  _TaxpayerRegimeCreate instance,
) => <String, dynamic>{
  'regime_code': instance.regimeCode,
  'options': ?instance.options,
  'start_period': ?instance.startPeriod,
};
