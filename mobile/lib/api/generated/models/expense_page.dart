// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'expense.dart';
import 'period_totals.dart';

part 'expense_page.freezed.dart';
part 'expense_page.g.dart';

@Freezed()
abstract class ExpensePage with _$ExpensePage {
  const factory ExpensePage({
    required List<Expense> items,
    required PeriodTotals totals,
    @JsonKey(name: 'next_cursor') String? nextCursor,
  }) = _ExpensePage;

  factory ExpensePage.fromJson(Map<String, Object?> json) =>
      _$ExpensePageFromJson(json);
}
