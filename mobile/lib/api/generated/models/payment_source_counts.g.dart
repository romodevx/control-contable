// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'payment_source_counts.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_PaymentSourceCounts _$PaymentSourceCountsFromJson(Map<String, dynamic> json) =>
    _PaymentSourceCounts(
      filedMonths: (json['filed_months'] as num).toInt(),
      taxInputMonths: (json['tax_input_months'] as num).toInt(),
      estimatedByAppMonths: (json['estimated_by_app_months'] as num).toInt(),
    );

Map<String, dynamic> _$PaymentSourceCountsToJson(
  _PaymentSourceCounts instance,
) => <String, dynamic>{
  'filed_months': instance.filedMonths,
  'tax_input_months': instance.taxInputMonths,
  'estimated_by_app_months': instance.estimatedByAppMonths,
};
