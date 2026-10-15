// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'salary_statement_list.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_SalaryStatementList _$SalaryStatementListFromJson(Map<String, dynamic> json) =>
    _SalaryStatementList(
      fiscalYear: (json['fiscal_year'] as num).toInt(),
      items: (json['items'] as List<dynamic>)
          .map((e) => SalaryStatement.fromJson(e as Map<String, dynamic>))
          .toList(),
      totals: SalaryStatementTotals.fromJson(
        json['totals'] as Map<String, dynamic>,
      ),
    );

Map<String, dynamic> _$SalaryStatementListToJson(
  _SalaryStatementList instance,
) => <String, dynamic>{
  'fiscal_year': instance.fiscalYear,
  'items': instance.items,
  'totals': instance.totals,
};
