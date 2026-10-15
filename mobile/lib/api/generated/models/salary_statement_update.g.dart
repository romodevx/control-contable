// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'salary_statement_update.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_SalaryStatementUpdate _$SalaryStatementUpdateFromJson(
  Map<String, dynamic> json,
) => _SalaryStatementUpdate(
  fiscalYear: (json['fiscal_year'] as num?)?.toInt(),
  employerLabel: json['employer_label'] as String?,
  accumulableCents: (json['accumulable_cents'] as num?)?.toInt(),
  exemptCents: (json['exempt_cents'] as num?)?.toInt(),
  isrWithheldCents: (json['isr_withheld_cents'] as num?)?.toInt(),
  simultaneous: json['simultaneous'] as bool?,
  leftBeforeYearEnd: json['left_before_year_end'] as bool?,
  payerNotWithholding: json['payer_not_withholding'] as bool?,
);

Map<String, dynamic> _$SalaryStatementUpdateToJson(
  _SalaryStatementUpdate instance,
) => <String, dynamic>{
  'fiscal_year': ?instance.fiscalYear,
  'employer_label': ?instance.employerLabel,
  'accumulable_cents': ?instance.accumulableCents,
  'exempt_cents': ?instance.exemptCents,
  'isr_withheld_cents': ?instance.isrWithheldCents,
  'simultaneous': ?instance.simultaneous,
  'left_before_year_end': ?instance.leftBeforeYearEnd,
  'payer_not_withholding': ?instance.payerNotWithholding,
};
