// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'deduction_category.dart';
import 'non_negative_cents.dart';

part 'deduction_category_total.freezed.dart';
part 'deduction_category_total.g.dart';

/// Registered amounts of one category (`excluded_cents`, the part that can never apply, is included in `registered_cents`).
@Freezed()
abstract class DeductionCategoryTotal with _$DeductionCategoryTotal {
  const factory DeductionCategoryTotal({
    required DeductionCategory category,
    @JsonKey(name: 'registered_cents')
    required NonNegativeCents registeredCents,
    @JsonKey(name: 'excluded_cents') required NonNegativeCents excludedCents,
    required int count,
  }) = _DeductionCategoryTotal;

  factory DeductionCategoryTotal.fromJson(Map<String, Object?> json) =>
      _$DeductionCategoryTotalFromJson(json);
}
