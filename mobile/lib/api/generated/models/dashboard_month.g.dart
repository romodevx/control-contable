// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'dashboard_month.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_DashboardMonth _$DashboardMonthFromJson(Map<String, dynamic> json) =>
    _DashboardMonth(
      incomeCents: (json['income_cents'] as num?)?.toInt(),
      expenseCents: (json['expense_cents'] as num?)?.toInt(),
    );

Map<String, dynamic> _$DashboardMonthToJson(_DashboardMonth instance) =>
    <String, dynamic>{
      'income_cents': instance.incomeCents,
      'expense_cents': instance.expenseCents,
    };
