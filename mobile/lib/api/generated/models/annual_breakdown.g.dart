// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'annual_breakdown.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_AnnualBreakdown _$AnnualBreakdownFromJson(Map<String, dynamic> json) =>
    _AnnualBreakdown(
      taxpayerId: json['taxpayer_id'] as String,
      fiscalYear: (json['fiscal_year'] as num).toInt(),
      section: AnnualSection.fromJson(json['section'] as Map<String, dynamic>),
      lines: (json['lines'] as List<dynamic>)
          .map((e) => AnnualBreakdownLine.fromJson(e as Map<String, dynamic>))
          .toList(),
      legend: json['legend'] as String,
      disclaimers: (json['disclaimers'] as List<dynamic>)
          .map((e) => Disclaimer.fromJson(e as Map<String, dynamic>))
          .toList(),
      engineVersion: json['engine_version'] as String,
    );

Map<String, dynamic> _$AnnualBreakdownToJson(_AnnualBreakdown instance) =>
    <String, dynamic>{
      'taxpayer_id': instance.taxpayerId,
      'fiscal_year': instance.fiscalYear,
      'section': instance.section,
      'lines': instance.lines,
      'legend': instance.legend,
      'disclaimers': instance.disclaimers,
      'engine_version': instance.engineVersion,
    };
