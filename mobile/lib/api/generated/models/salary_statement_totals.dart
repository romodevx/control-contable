// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'non_negative_cents.dart';

part 'salary_statement_totals.freezed.dart';
part 'salary_statement_totals.g.dart';

@Freezed()
abstract class SalaryStatementTotals with _$SalaryStatementTotals {
  const factory SalaryStatementTotals({
    @JsonKey(name: 'accumulable_cents')
    required NonNegativeCents accumulableCents,
    @JsonKey(name: 'exempt_cents') required NonNegativeCents exemptCents,
    @JsonKey(name: 'isr_withheld_cents')
    required NonNegativeCents isrWithheldCents,
  }) = _SalaryStatementTotals;

  factory SalaryStatementTotals.fromJson(Map<String, Object?> json) =>
      _$SalaryStatementTotalsFromJson(json);
}
