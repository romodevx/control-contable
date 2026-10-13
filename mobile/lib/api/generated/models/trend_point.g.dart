// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'trend_point.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_TrendPoint _$TrendPointFromJson(Map<String, dynamic> json) => _TrendPoint(
  period: json['period'] as String,
  incomeCents: (json['income_cents'] as num).toInt(),
  expenseCents: (json['expense_cents'] as num).toInt(),
  taxPayableRoundedCents: (json['tax_payable_rounded_cents'] as num).toInt(),
);

Map<String, dynamic> _$TrendPointToJson(_TrendPoint instance) =>
    <String, dynamic>{
      'period': instance.period,
      'income_cents': instance.incomeCents,
      'expense_cents': instance.expenseCents,
      'tax_payable_rounded_cents': instance.taxPayableRoundedCents,
    };
