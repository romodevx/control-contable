// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'fiscal_year.dart';
import 'salary_statement.dart';
import 'salary_statement_totals.dart';

part 'salary_statement_list.freezed.dart';
part 'salary_statement_list.g.dart';

@Freezed()
abstract class SalaryStatementList with _$SalaryStatementList {
  const factory SalaryStatementList({
    @JsonKey(name: 'fiscal_year') required FiscalYear fiscalYear,
    required List<SalaryStatement> items,
    required SalaryStatementTotals totals,
  }) = _SalaryStatementList;

  factory SalaryStatementList.fromJson(Map<String, Object?> json) =>
      _$SalaryStatementListFromJson(json);
}
