// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'deductibility.dart';

part 'expense_category.freezed.dart';
part 'expense_category.g.dart';

@Freezed()
abstract class ExpenseCategory with _$ExpenseCategory {
  const factory ExpenseCategory({
    required String code,

    /// es-MX display name.
    required String name,
    required Deductibility deductibility,

    /// The `deductible` a new expense starts with (before the CFDI and cash rules); conservative for partial cases.
    @JsonKey(name: 'default_deductible') required bool defaultDeductible,

    /// es-MX note shown with the hint.
    @JsonKey(includeIfNull: true) required String? note,

    /// Article-level basis (e.g. `LISR 103`).
    @JsonKey(includeIfNull: true, name: 'legal_basis')
    required String? legalBasis,
  }) = _ExpenseCategory;

  factory ExpenseCategory.fromJson(Map<String, Object?> json) =>
      _$ExpenseCategoryFromJson(json);
}
