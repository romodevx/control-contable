// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'expense_category.dart';
import 'legend.dart';

part 'expense_category_list.freezed.dart';
part 'expense_category_list.g.dart';

@Freezed()
abstract class ExpenseCategoryList with _$ExpenseCategoryList {
  const factory ExpenseCategoryList({
    required List<ExpenseCategory> items,
    required Legend legend,
  }) = _ExpenseCategoryList;

  factory ExpenseCategoryList.fromJson(Map<String, Object?> json) =>
      _$ExpenseCategoryListFromJson(json);
}
