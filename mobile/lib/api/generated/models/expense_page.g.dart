// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'expense_page.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_ExpensePage _$ExpensePageFromJson(Map<String, dynamic> json) => _ExpensePage(
  items: (json['items'] as List<dynamic>)
      .map((e) => Expense.fromJson(e as Map<String, dynamic>))
      .toList(),
  nextCursor: json['next_cursor'] as String?,
  totals: PeriodTotals.fromJson(json['totals'] as Map<String, dynamic>),
  legend: json['legend'] as String,
);

Map<String, dynamic> _$ExpensePageToJson(_ExpensePage instance) =>
    <String, dynamic>{
      'items': instance.items,
      'next_cursor': instance.nextCursor,
      'totals': instance.totals,
      'legend': instance.legend,
    };
