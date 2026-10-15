// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'taxpayer_regime.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_TaxpayerRegime _$TaxpayerRegimeFromJson(Map<String, dynamic> json) =>
    _TaxpayerRegime(
      regimeCode: json['regime_code'] as String,
      name: json['name'] as String,
      treatment: Treatment.fromJson(json['treatment'] as String),
      incomeBasis: IncomeBasis.fromJson(json['income_basis'] as String),
      options: json['options'] as Map<String, dynamic>,
      startPeriod: json['start_period'] as String,
      endPeriod: json['end_period'] as String?,
      warnings: (json['warnings'] as List<dynamic>?)
          ?.map((e) => Warning.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$TaxpayerRegimeToJson(_TaxpayerRegime instance) =>
    <String, dynamic>{
      'regime_code': instance.regimeCode,
      'name': instance.name,
      'treatment': _$TreatmentEnumMap[instance.treatment]!,
      'income_basis': _$IncomeBasisEnumMap[instance.incomeBasis]!,
      'options': instance.options,
      'start_period': instance.startPeriod,
      'end_period': instance.endPeriod,
      'warnings': ?instance.warnings,
    };

const _$TreatmentEnumMap = {
  Treatment.full: 'full',
  Treatment.partial: 'partial',
  Treatment.informational: 'informational',
  Treatment.$unknown: r'$unknown',
};

const _$IncomeBasisEnumMap = {
  IncomeBasis.cash: 'cash',
  IncomeBasis.accrual: 'accrual',
  IncomeBasis.$unknown: r'$unknown',
};
