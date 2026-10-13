// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'access_summary.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_AccessSummary _$AccessSummaryFromJson(Map<String, dynamic> json) =>
    _AccessSummary(
      access: AccessSummaryAccess.fromJson(json['access'] as String),
      trialEndsAt: json['trial_ends_at'] == null
          ? null
          : DateTime.parse(json['trial_ends_at'] as String),
      incomesRemaining: (json['incomes_remaining'] as num?)?.toInt(),
      expensesRemaining: (json['expenses_remaining'] as num?)?.toInt(),
    );

Map<String, dynamic> _$AccessSummaryToJson(_AccessSummary instance) =>
    <String, dynamic>{
      'access': _$AccessSummaryAccessEnumMap[instance.access]!,
      'trial_ends_at': ?instance.trialEndsAt?.toIso8601String(),
      'incomes_remaining': ?instance.incomesRemaining,
      'expenses_remaining': ?instance.expensesRemaining,
    };

const _$AccessSummaryAccessEnumMap = {
  AccessSummaryAccess.trial: 'trial',
  AccessSummaryAccess.pro: 'pro',
  AccessSummaryAccess.comp: 'comp',
  AccessSummaryAccess.readOnly: 'read_only',
  AccessSummaryAccess.$unknown: r'$unknown',
};
