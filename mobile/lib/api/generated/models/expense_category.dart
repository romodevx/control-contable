// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'deductibility.dart';
import 'operation_kind.dart';

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

    /// The `operation_kind` of this category (seed data in `category_regime_rules`, MODULE-withholdings §2.2): what a persona moral's "Tu proveedor" question E3 starts at, and what `createExpense`, `updateExpense` and `previewExpenseWithholdings` use when `operation_kind` is omitted. Clients never keep their own copy of this table. Optional in 0.11.1 (additive); sent once WH-03 seeds it and WH-07 serves it, when it **will be required**. Absent, a client leaves E3 unselected and omits `operation_kind`, so the server applies the same default.
    @JsonKey(includeIfNull: false, name: 'default_operation_kind')
    OperationKind? defaultOperationKind,
  }) = _ExpenseCategory;

  factory ExpenseCategory.fromJson(Map<String, Object?> json) =>
      _$ExpenseCategoryFromJson(json);
}
