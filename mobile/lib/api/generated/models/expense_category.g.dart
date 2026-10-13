// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'expense_category.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_ExpenseCategory _$ExpenseCategoryFromJson(Map<String, dynamic> json) =>
    _ExpenseCategory(
      code: json['code'] as String,
      name: json['name'] as String,
      hints: (json['hints'] as List<dynamic>)
          .map((e) => DeductibilityHint.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$ExpenseCategoryToJson(_ExpenseCategory instance) =>
    <String, dynamic>{
      'code': instance.code,
      'name': instance.name,
      'hints': instance.hints,
    };
