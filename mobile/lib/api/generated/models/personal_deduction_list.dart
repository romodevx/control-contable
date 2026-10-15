// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'deduction_category_total.dart';
import 'fiscal_year.dart';
import 'personal_deduction.dart';

part 'personal_deduction_list.freezed.dart';
part 'personal_deduction_list.g.dart';

/// `totals` has one entry per category with at least one row, in `DeductionCategory` order.
@Freezed()
abstract class PersonalDeductionList with _$PersonalDeductionList {
  const factory PersonalDeductionList({
    @JsonKey(name: 'fiscal_year') required FiscalYear fiscalYear,
    required List<PersonalDeduction> items,
    required List<DeductionCategoryTotal> totals,
  }) = _PersonalDeductionList;

  factory PersonalDeductionList.fromJson(Map<String, Object?> json) =>
      _$PersonalDeductionListFromJson(json);
}
