// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'dashboard.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Dashboard _$DashboardFromJson(Map<String, dynamic> json) => _Dashboard(
  taxpayerId: json['taxpayer_id'] as String,
  period: json['period'] as String,
  isEmpty: json['is_empty'] as bool,
  estimate: TaxpayerEstimate.fromJson(json['estimate'] as Map<String, dynamic>),
  incomeCents: (json['income_cents'] as num).toInt(),
  expenseCents: (json['expense_cents'] as num).toInt(),
  trend: (json['trend'] as List<dynamic>)
      .map((e) => TrendPoint.fromJson(e as Map<String, dynamic>))
      .toList(),
  access: AccessSummary.fromJson(json['access'] as Map<String, dynamic>),
  legend: json['legend'] as String,
);

Map<String, dynamic> _$DashboardToJson(_Dashboard instance) =>
    <String, dynamic>{
      'taxpayer_id': instance.taxpayerId,
      'period': instance.period,
      'is_empty': instance.isEmpty,
      'estimate': instance.estimate,
      'income_cents': instance.incomeCents,
      'expense_cents': instance.expenseCents,
      'trend': instance.trend,
      'access': instance.access,
      'legend': instance.legend,
    };
