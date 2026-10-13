// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'history_year.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_HistoryYear _$HistoryYearFromJson(Map<String, dynamic> json) => _HistoryYear(
  taxpayerId: json['taxpayer_id'] as String,
  fiscalYear: (json['fiscal_year'] as num).toInt(),
  months: (json['months'] as List<dynamic>)
      .map((e) => HistoryMonth.fromJson(e as Map<String, dynamic>))
      .toList(),
  legend: json['legend'] as String,
);

Map<String, dynamic> _$HistoryYearToJson(_HistoryYear instance) =>
    <String, dynamic>{
      'taxpayer_id': instance.taxpayerId,
      'fiscal_year': instance.fiscalYear,
      'months': instance.months,
      'legend': instance.legend,
    };
