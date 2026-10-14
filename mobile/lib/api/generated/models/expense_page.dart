// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'expense.dart';
import 'legend.dart';
import 'period_totals.dart';

part 'expense_page.freezed.dart';
part 'expense_page.g.dart';

@Freezed()
abstract class ExpensePage with _$ExpensePage {
  const factory ExpensePage({
    required List<Expense> items,

    /// Opaque cursor of the next page; `null` on the last page.
    @JsonKey(includeIfNull: true, name: 'next_cursor')
    required String? nextCursor,
    required PeriodTotals totals,
    required Legend legend,
  }) = _ExpensePage;

  factory ExpensePage.fromJson(Map<String, Object?> json) =>
      _$ExpensePageFromJson(json);
}
