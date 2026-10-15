// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'employer_label.dart';
import 'fiscal_year.dart';
import 'non_negative_cents.dart';

part 'salary_statement_update.freezed.dart';
part 'salary_statement_update.g.dart';

/// Merge-patch; `employer_label: null` clears it.
@Freezed()
abstract class SalaryStatementUpdate with _$SalaryStatementUpdate {
  const factory SalaryStatementUpdate({
    @JsonKey(includeIfNull: false, name: 'fiscal_year') FiscalYear? fiscalYear,
    @JsonKey(includeIfNull: false, name: 'employer_label')
    EmployerLabel? employerLabel,
    @JsonKey(includeIfNull: false, name: 'accumulable_cents')
    NonNegativeCents? accumulableCents,
    @JsonKey(includeIfNull: false, name: 'exempt_cents')
    NonNegativeCents? exemptCents,
    @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')
    NonNegativeCents? isrWithheldCents,
    @JsonKey(includeIfNull: false) bool? simultaneous,
    @JsonKey(includeIfNull: false, name: 'left_before_year_end')
    bool? leftBeforeYearEnd,
    @JsonKey(includeIfNull: false, name: 'payer_not_withholding')
    bool? payerNotWithholding,
  }) = _SalaryStatementUpdate;

  factory SalaryStatementUpdate.fromJson(Map<String, Object?> json) =>
      _$SalaryStatementUpdateFromJson(json);
}
