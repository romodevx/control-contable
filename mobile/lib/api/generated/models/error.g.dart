// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'error.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Error _$ErrorFromJson(Map<String, dynamic> json) => _Error(
  code: ErrorCode.fromJson(json['code'] as String),
  message: json['message'] as String,
  requestId: json['request_id'] as String,
  details: (json['details'] as List<dynamic>?)
      ?.map((e) => FieldError.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$ErrorToJson(_Error instance) => <String, dynamic>{
  'code': _$ErrorCodeEnumMap[instance.code]!,
  'message': instance.message,
  'request_id': instance.requestId,
  'details': ?instance.details,
};

const _$ErrorCodeEnumMap = {
  ErrorCode.validationFailed: 'validation_failed',
  ErrorCode.unauthorized: 'unauthorized',
  ErrorCode.entitlementRequired: 'entitlement_required',
  ErrorCode.notFound: 'not_found',
  ErrorCode.conflict: 'conflict',
  ErrorCode.rfcAlreadyRegistered: 'rfc_already_registered',
  ErrorCode.regimeHasRecords: 'regime_has_records',
  ErrorCode.invalidRegimeCombination: 'invalid_regime_combination',
  ErrorCode.businessRuleViolation: 'business_rule_violation',
  ErrorCode.taxpayerLimitReached: 'taxpayer_limit_reached',
  ErrorCode.rateLimited: 'rate_limited',
  ErrorCode.upstreamUnavailable: 'upstream_unavailable',
  ErrorCode.internalError: 'internal_error',
  ErrorCode.invalidPassword: 'invalid_password',
  ErrorCode.$unknown: r'$unknown',
};
