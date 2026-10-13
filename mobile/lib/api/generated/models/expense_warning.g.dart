// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'expense_warning.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_ExpenseWarning _$ExpenseWarningFromJson(Map<String, dynamic> json) =>
    _ExpenseWarning(
      code: ExpenseWarningCode.fromJson(json['code'] as String),
      message: json['message'] as String,
    );

Map<String, dynamic> _$ExpenseWarningToJson(_ExpenseWarning instance) =>
    <String, dynamic>{
      'code': _$ExpenseWarningCodeEnumMap[instance.code]!,
      'message': instance.message,
    };

const _$ExpenseWarningCodeEnumMap = {
  ExpenseWarningCode.cashOverThreshold: 'cash_over_threshold',
  ExpenseWarningCode.noCfdi: 'no_cfdi',
  ExpenseWarningCode.notIsrDeductibleInRegime: 'not_isr_deductible_in_regime',
  ExpenseWarningCode.categoryNotDeductible: 'category_not_deductible',
  ExpenseWarningCode.$unknown: r'$unknown',
};
