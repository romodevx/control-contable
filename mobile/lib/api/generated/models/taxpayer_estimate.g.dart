// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'taxpayer_estimate.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_TaxpayerEstimate _$TaxpayerEstimateFromJson(Map<String, dynamic> json) =>
    _TaxpayerEstimate(
      taxpayerId: json['taxpayer_id'] as String,
      period: json['period'] as String,
      dueDate: json['due_date'] as String,
      legend: json['legend'] as String,
      disclaimers: (json['disclaimers'] as List<dynamic>)
          .map((e) => Disclaimer.fromJson(e as Map<String, dynamic>))
          .toList(),
      regimes: (json['regimes'] as List<dynamic>)
          .map((e) => RegimeEstimate.fromJson(e as Map<String, dynamic>))
          .toList(),
      totals: EstimateTotals.fromJson(json['totals'] as Map<String, dynamic>),
      parameterSet: ParameterSetRef.fromJson(
        json['parameter_set'] as Map<String, dynamic>,
      ),
      recordsFingerprint: json['records_fingerprint'] as String,
      engineVersion: json['engine_version'] as String,
    );

Map<String, dynamic> _$TaxpayerEstimateToJson(_TaxpayerEstimate instance) =>
    <String, dynamic>{
      'taxpayer_id': instance.taxpayerId,
      'period': instance.period,
      'due_date': instance.dueDate,
      'legend': instance.legend,
      'disclaimers': instance.disclaimers,
      'regimes': instance.regimes,
      'totals': instance.totals,
      'parameter_set': instance.parameterSet,
      'records_fingerprint': instance.recordsFingerprint,
      'engine_version': instance.engineVersion,
    };
