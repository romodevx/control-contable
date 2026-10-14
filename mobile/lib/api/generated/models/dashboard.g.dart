// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'dashboard.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Dashboard _$DashboardFromJson(Map<String, dynamic> json) => _Dashboard(
  taxpayerId: json['taxpayer_id'] as String,
  period: json['period'] as String,
  isEmpty: json['is_empty'] as bool,
  month: DashboardMonth.fromJson(json['month'] as Map<String, dynamic>),
  estimate: json['estimate'] == null
      ? null
      : TaxpayerEstimate.fromJson(json['estimate'] as Map<String, dynamic>),
  trend: (json['trend'] as List<dynamic>)
      .map((e) => TrendPoint.fromJson(e as Map<String, dynamic>))
      .toList(),
  access: AccessSummary.fromJson(json['access'] as Map<String, dynamic>),
  legend: json['legend'] as String?,
);

Map<String, dynamic> _$DashboardToJson(_Dashboard instance) =>
    <String, dynamic>{
      'taxpayer_id': instance.taxpayerId,
      'period': instance.period,
      'is_empty': instance.isEmpty,
      'month': instance.month,
      'estimate': instance.estimate,
      'trend': instance.trend,
      'access': instance.access,
      'legend': ?instance.legend,
    };
