// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'employer_label.dart';
import 'fiscal_year.dart';
import 'non_negative_cents.dart';
import 'uuid.dart';

part 'salary_statement_create.freezed.dart';
part 'salary_statement_create.g.dart';

/// The three booleans default to `false`; `id` is the client's idempotency key.
@Freezed()
abstract class SalaryStatementCreate with _$SalaryStatementCreate {
  const factory SalaryStatementCreate({
    @JsonKey(name: 'fiscal_year') required FiscalYear fiscalYear,
    @JsonKey(name: 'accumulable_cents')
    required NonNegativeCents accumulableCents,
    @JsonKey(name: 'exempt_cents') required NonNegativeCents exemptCents,
    @JsonKey(name: 'isr_withheld_cents')
    required NonNegativeCents isrWithheldCents,
    @Default(false) bool simultaneous,
    @JsonKey(name: 'left_before_year_end')
    @Default(false)
    bool leftBeforeYearEnd,
    @JsonKey(name: 'payer_not_withholding')
    @Default(false)
    bool payerNotWithholding,
    @JsonKey(includeIfNull: false) Uuid? id,
    @JsonKey(includeIfNull: false, name: 'employer_label')
    EmployerLabel? employerLabel,
  }) = _SalaryStatementCreate;

  factory SalaryStatementCreate.fromJson(Map<String, Object?> json) =>
      _$SalaryStatementCreateFromJson(json);
}
