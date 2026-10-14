// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'history_viewed_parameter_set.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_HistoryViewedParameterSet _$HistoryViewedParameterSetFromJson(
  Map<String, dynamic> json,
) => _HistoryViewedParameterSet(
  fiscalYear: (json['fiscal_year'] as num).toInt(),
  version: (json['version'] as num).toInt(),
);

Map<String, dynamic> _$HistoryViewedParameterSetToJson(
  _HistoryViewedParameterSet instance,
) => <String, dynamic>{
  'fiscal_year': instance.fiscalYear,
  'version': instance.version,
};
