// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'history_month.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_HistoryMonth _$HistoryMonthFromJson(Map<String, dynamic> json) =>
    _HistoryMonth(
      period: json['period'] as String,
      status: TaxpayerEstimateStatus.fromJson(json['status'] as String),
      hasRecords: json['has_records'] as bool,
      tax: json['tax'] == null
          ? null
          : HistoryTax.fromJson(json['tax'] as Map<String, dynamic>),
      dueDate: json['due_date'] as String,
      cpaReviewPending: json['cpa_review_pending'] as bool,
      isCurrent: json['is_current'] as bool,
      recordsFingerprint: json['records_fingerprint'] as String,
      figuresFingerprint: json['figures_fingerprint'] as String,
      engineVersion: json['engine_version'] as String,
      parameterSet: ParameterSetRef.fromJson(
        json['parameter_set'] as Map<String, dynamic>,
      ),
      viewedAt: json['viewed_at'] == null
          ? null
          : DateTime.parse(json['viewed_at'] as String),
      changedSinceViewed: json['changed_since_viewed'] as bool,
      changeReason: json['change_reason'] == null
          ? null
          : HistoryChangeReason.fromJson(json['change_reason'] as String),
    );

Map<String, dynamic> _$HistoryMonthToJson(_HistoryMonth instance) =>
    <String, dynamic>{
      'period': instance.period,
      'status': _$TaxpayerEstimateStatusEnumMap[instance.status]!,
      'has_records': instance.hasRecords,
      'tax': instance.tax,
      'due_date': instance.dueDate,
      'cpa_review_pending': instance.cpaReviewPending,
      'is_current': instance.isCurrent,
      'records_fingerprint': instance.recordsFingerprint,
      'figures_fingerprint': instance.figuresFingerprint,
      'engine_version': instance.engineVersion,
      'parameter_set': instance.parameterSet,
      'viewed_at': instance.viewedAt?.toIso8601String(),
      'changed_since_viewed': instance.changedSinceViewed,
      'change_reason': _$HistoryChangeReasonEnumMap[instance.changeReason],
    };

const _$TaxpayerEstimateStatusEnumMap = {
  TaxpayerEstimateStatus.complete: 'complete',
  TaxpayerEstimateStatus.incomplete: 'incomplete',
  TaxpayerEstimateStatus.partial: 'partial',
  TaxpayerEstimateStatus.informational: 'informational',
  TaxpayerEstimateStatus.none: 'none',
  TaxpayerEstimateStatus.$unknown: r'$unknown',
};

const _$HistoryChangeReasonEnumMap = {
  HistoryChangeReason.records: 'records',
  HistoryChangeReason.engine: 'engine',
  HistoryChangeReason.$unknown: r'$unknown',
};
