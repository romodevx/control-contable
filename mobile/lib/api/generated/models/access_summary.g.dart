// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'access_summary.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_AccessSummary _$AccessSummaryFromJson(Map<String, dynamic> json) =>
    _AccessSummary(
      access: AccessLevel.fromJson(json['access'] as String),
      trialEndsAt: json['trial_ends_at'] == null
          ? null
          : DateTime.parse(json['trial_ends_at'] as String),
      trialDaysLeft: (json['trial_days_left'] as num?)?.toInt(),
      incomesRemaining: (json['incomes_remaining'] as num?)?.toInt(),
      expensesRemaining: (json['expenses_remaining'] as num?)?.toInt(),
    );

Map<String, dynamic> _$AccessSummaryToJson(_AccessSummary instance) =>
    <String, dynamic>{
      'access': _$AccessLevelEnumMap[instance.access]!,
      'trial_ends_at': instance.trialEndsAt?.toIso8601String(),
      'trial_days_left': instance.trialDaysLeft,
      'incomes_remaining': instance.incomesRemaining,
      'expenses_remaining': instance.expensesRemaining,
    };

const _$AccessLevelEnumMap = {
  AccessLevel.trial: 'trial',
  AccessLevel.pro: 'pro',
  AccessLevel.comp: 'comp',
  AccessLevel.readOnly: 'read_only',
  AccessLevel.$unknown: r'$unknown',
};
