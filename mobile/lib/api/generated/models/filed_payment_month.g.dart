// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'filed_payment_month.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_FiledPaymentMonth _$FiledPaymentMonthFromJson(Map<String, dynamic> json) =>
    _FiledPaymentMonth(
      month: (json['month'] as num).toInt(),
      period: json['period'] as String,
      inWindow: json['in_window'] as bool,
      filedCents: (json['filed_cents'] as num?)?.toInt(),
      taxInputCents: (json['tax_input_cents'] as num?)?.toInt(),
      appEstimateCents: (json['app_estimate_cents'] as num?)?.toInt(),
      appEstimateIncomplete: json['app_estimate_incomplete'] as bool,
      effectiveCents: (json['effective_cents'] as num?)?.toInt(),
      source: FiledPaymentSource.fromJson(json['source'] as String),
    );

Map<String, dynamic> _$FiledPaymentMonthToJson(_FiledPaymentMonth instance) =>
    <String, dynamic>{
      'month': instance.month,
      'period': instance.period,
      'in_window': instance.inWindow,
      'filed_cents': instance.filedCents,
      'tax_input_cents': instance.taxInputCents,
      'app_estimate_cents': instance.appEstimateCents,
      'app_estimate_incomplete': instance.appEstimateIncomplete,
      'effective_cents': instance.effectiveCents,
      'source': _$FiledPaymentSourceEnumMap[instance.source]!,
    };

const _$FiledPaymentSourceEnumMap = {
  FiledPaymentSource.filed: 'filed',
  FiledPaymentSource.taxInput: 'tax_input',
  FiledPaymentSource.estimatedByApp: 'estimated_by_app',
  FiledPaymentSource.none: 'none',
  FiledPaymentSource.$unknown: r'$unknown',
};
