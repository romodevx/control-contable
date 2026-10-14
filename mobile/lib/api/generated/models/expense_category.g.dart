// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'expense_category.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_ExpenseCategory _$ExpenseCategoryFromJson(Map<String, dynamic> json) =>
    _ExpenseCategory(
      code: json['code'] as String,
      name: json['name'] as String,
      deductibility: Deductibility.fromJson(json['deductibility'] as String),
      defaultDeductible: json['default_deductible'] as bool,
      note: json['note'] as String?,
      legalBasis: json['legal_basis'] as String?,
    );

Map<String, dynamic> _$ExpenseCategoryToJson(_ExpenseCategory instance) =>
    <String, dynamic>{
      'code': instance.code,
      'name': instance.name,
      'deductibility': _$DeductibilityEnumMap[instance.deductibility]!,
      'default_deductible': instance.defaultDeductible,
      'note': instance.note,
      'legal_basis': instance.legalBasis,
    };

const _$DeductibilityEnumMap = {
  Deductibility.deductible: 'deductible',
  Deductibility.partiallyDeductible: 'partially_deductible',
  Deductibility.notDeductible: 'not_deductible',
  Deductibility.isrNotApplicable: 'isr_not_applicable',
  Deductibility.$unknown: r'$unknown',
};
