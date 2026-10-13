// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'deductibility_hint.dart';

part 'expense_category.freezed.dart';
part 'expense_category.g.dart';

@Freezed()
abstract class ExpenseCategory with _$ExpenseCategory {
  const factory ExpenseCategory({
    required String code,
    required String name,
    required List<DeductibilityHint> hints,
  }) = _ExpenseCategory;

  factory ExpenseCategory.fromJson(Map<String, Object?> json) =>
      _$ExpenseCategoryFromJson(json);
}
