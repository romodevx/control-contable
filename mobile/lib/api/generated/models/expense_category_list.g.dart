// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'expense_category_list.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_ExpenseCategoryList _$ExpenseCategoryListFromJson(Map<String, dynamic> json) =>
    _ExpenseCategoryList(
      items: (json['items'] as List<dynamic>)
          .map((e) => ExpenseCategory.fromJson(e as Map<String, dynamic>))
          .toList(),
      legend: json['legend'] as String,
    );

Map<String, dynamic> _$ExpenseCategoryListToJson(
  _ExpenseCategoryList instance,
) => <String, dynamic>{'items': instance.items, 'legend': instance.legend};
