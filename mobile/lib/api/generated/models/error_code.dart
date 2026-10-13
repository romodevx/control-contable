// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

@JsonEnum()
enum ErrorCode {
  @JsonValue('validation_failed')
  validationFailed('validation_failed'),
  @JsonValue('unauthorized')
  unauthorized('unauthorized'),
  @JsonValue('entitlement_required')
  entitlementRequired('entitlement_required'),
  @JsonValue('not_found')
  notFound('not_found'),
  @JsonValue('conflict')
  conflict('conflict'),
  @JsonValue('rfc_already_registered')
  rfcAlreadyRegistered('rfc_already_registered'),
  @JsonValue('regime_has_records')
  regimeHasRecords('regime_has_records'),
  @JsonValue('invalid_regime_combination')
  invalidRegimeCombination('invalid_regime_combination'),
  @JsonValue('business_rule_violation')
  businessRuleViolation('business_rule_violation'),
  @JsonValue('taxpayer_limit_reached')
  taxpayerLimitReached('taxpayer_limit_reached'),
  @JsonValue('rate_limited')
  rateLimited('rate_limited'),
  @JsonValue('upstream_unavailable')
  upstreamUnavailable('upstream_unavailable'),
  @JsonValue('internal_error')
  internalError('internal_error'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const ErrorCode(this.json);

  factory ErrorCode.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<ErrorCode> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
