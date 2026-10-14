// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'history_viewed_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_HistoryViewedRequest _$HistoryViewedRequestFromJson(
  Map<String, dynamic> json,
) => _HistoryViewedRequest(
  recordsFingerprint: json['records_fingerprint'] as String,
  figuresFingerprint: json['figures_fingerprint'] as String,
  engineVersion: json['engine_version'] as String,
  parameterSet: HistoryViewedParameterSet.fromJson(
    json['parameter_set'] as Map<String, dynamic>,
  ),
);

Map<String, dynamic> _$HistoryViewedRequestToJson(
  _HistoryViewedRequest instance,
) => <String, dynamic>{
  'records_fingerprint': instance.recordsFingerprint,
  'figures_fingerprint': instance.figuresFingerprint,
  'engine_version': instance.engineVersion,
  'parameter_set': instance.parameterSet,
};
