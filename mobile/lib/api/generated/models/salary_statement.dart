// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'employer_label.dart';
import 'fiscal_year.dart';
import 'non_negative_cents.dart';
import 'uuid.dart';
import 'warning.dart';

part 'salary_statement.freezed.dart';
part 'salary_statement.g.dart';

/// One employer's constancia de sueldos (605) totals, as printed: the app never recomputes exemptions or the employment subsidy. The three flags feed the obligation test (LISR 98 fr. III). `warnings` is filled on writes only.
@Freezed()
abstract class SalaryStatement with _$SalaryStatement {
  const factory SalaryStatement({
    required Uuid id,
    @JsonKey(name: 'taxpayer_id') required Uuid taxpayerId,
    @JsonKey(name: 'fiscal_year') required FiscalYear fiscalYear,
    @JsonKey(includeIfNull: true, name: 'employer_label')
    required EmployerLabel? employerLabel,
    @JsonKey(name: 'accumulable_cents')
    required NonNegativeCents accumulableCents,
    @JsonKey(name: 'exempt_cents') required NonNegativeCents exemptCents,
    @JsonKey(name: 'isr_withheld_cents')
    required NonNegativeCents isrWithheldCents,

    /// Worked for two or more employers at the same time.
    required bool simultaneous,
    @JsonKey(name: 'left_before_year_end') required bool leftBeforeYearEnd,

    /// The employer was not obliged to withhold (e.g. an international organization).
    @JsonKey(name: 'payer_not_withholding') required bool payerNotWithholding,
    required List<Warning> warnings,
    @JsonKey(name: 'created_at') required DateTime createdAt,
    @JsonKey(name: 'updated_at') required DateTime updatedAt,
  }) = _SalaryStatement;

  factory SalaryStatement.fromJson(Map<String, Object?> json) =>
      _$SalaryStatementFromJson(json);
}
