// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'health_status.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_HealthStatus _$HealthStatusFromJson(Map<String, dynamic> json) =>
    _HealthStatus(
      status: HealthStatusStatus.fromJson(json['status'] as String),
      database: HealthStatusDatabase.fromJson(json['database'] as String),
      version: json['version'] as String,
    );

Map<String, dynamic> _$HealthStatusToJson(_HealthStatus instance) =>
    <String, dynamic>{
      'status': _$HealthStatusStatusEnumMap[instance.status]!,
      'database': _$HealthStatusDatabaseEnumMap[instance.database]!,
      'version': instance.version,
    };

const _$HealthStatusStatusEnumMap = {
  HealthStatusStatus.ok: 'ok',
  HealthStatusStatus.degraded: 'degraded',
  HealthStatusStatus.$unknown: r'$unknown',
};

const _$HealthStatusDatabaseEnumMap = {
  HealthStatusDatabase.ok: 'ok',
  HealthStatusDatabase.unreachable: 'unreachable',
  HealthStatusDatabase.$unknown: r'$unknown',
};
