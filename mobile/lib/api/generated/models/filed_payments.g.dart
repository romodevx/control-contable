// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'filed_payments.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_FiledPayments _$FiledPaymentsFromJson(Map<String, dynamic> json) =>
    _FiledPayments(
      regimeCode: json['regime_code'] as String,
      fiscalYear: (json['fiscal_year'] as num).toInt(),
      months: (json['months'] as List<dynamic>)
          .map((e) => FiledPaymentMonth.fromJson(e as Map<String, dynamic>))
          .toList(),
      totalEffectiveCents: (json['total_effective_cents'] as num).toInt(),
      updatedAt: json['updated_at'] == null
          ? null
          : DateTime.parse(json['updated_at'] as String),
    );

Map<String, dynamic> _$FiledPaymentsToJson(_FiledPayments instance) =>
    <String, dynamic>{
      'regime_code': instance.regimeCode,
      'fiscal_year': instance.fiscalYear,
      'months': instance.months,
      'total_effective_cents': instance.totalEffectiveCents,
      'updated_at': instance.updatedAt?.toIso8601String(),
    };
