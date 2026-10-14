// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'expense_category.dart';
import 'iva_rate.dart';
import 'legend.dart';
import 'non_negative_cents.dart';

part 'expense_category_list.freezed.dart';
part 'expense_category_list.g.dart';

@Freezed()
abstract class ExpenseCategoryList with _$ExpenseCategoryList {
  const factory ExpenseCategoryList({
    required List<ExpenseCategory> items,
    @JsonKey(name: 'iva_rates_allowed') required List<IvaRate> ivaRatesAllowed,

    /// A cash payment above this total (with IVA) is not deductible (LISR 27-III; dated parameter data, ⚠️ CPA).
    @JsonKey(name: 'cash_payment_limit_cents')
    required NonNegativeCents cashPaymentLimitCents,
    required Legend legend,
  }) = _ExpenseCategoryList;

  factory ExpenseCategoryList.fromJson(Map<String, Object?> json) =>
      _$ExpenseCategoryListFromJson(json);
}
