// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'history_month.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_HistoryMonth _$HistoryMonthFromJson(Map<String, dynamic> json) =>
    _HistoryMonth(
      period: json['period'] as String,
      status: HistoryMonthStatus.fromJson(json['status'] as String),
      totalPayableRoundedCents: (json['total_payable_rounded_cents'] as num)
          .toInt(),
      changedSinceViewed: json['changed_since_viewed'] as bool,
      iva: json['iva'] == null
          ? null
          : TaxResult.fromJson(json['iva'] as Map<String, dynamic>),
      isr: json['isr'] == null
          ? null
          : TaxResult.fromJson(json['isr'] as Map<String, dynamic>),
      lastViewedAt: json['last_viewed_at'] == null
          ? null
          : DateTime.parse(json['last_viewed_at'] as String),
    );

Map<String, dynamic> _$HistoryMonthToJson(_HistoryMonth instance) =>
    <String, dynamic>{
      'period': instance.period,
      'status': _$HistoryMonthStatusEnumMap[instance.status]!,
      'total_payable_rounded_cents': instance.totalPayableRoundedCents,
      'changed_since_viewed': instance.changedSinceViewed,
      'iva': instance.iva,
      'isr': instance.isr,
      'last_viewed_at': instance.lastViewedAt?.toIso8601String(),
    };

const _$HistoryMonthStatusEnumMap = {
  HistoryMonthStatus.complete: 'complete',
  HistoryMonthStatus.incomplete: 'incomplete',
  HistoryMonthStatus.informational: 'informational',
  HistoryMonthStatus.noRecords: 'no_records',
  HistoryMonthStatus.$unknown: r'$unknown',
};
