// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'salary_statement_create.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_SalaryStatementCreate _$SalaryStatementCreateFromJson(
  Map<String, dynamic> json,
) => _SalaryStatementCreate(
  fiscalYear: (json['fiscal_year'] as num).toInt(),
  accumulableCents: (json['accumulable_cents'] as num).toInt(),
  exemptCents: (json['exempt_cents'] as num).toInt(),
  isrWithheldCents: (json['isr_withheld_cents'] as num).toInt(),
  simultaneous: json['simultaneous'] as bool? ?? false,
  leftBeforeYearEnd: json['left_before_year_end'] as bool? ?? false,
  payerNotWithholding: json['payer_not_withholding'] as bool? ?? false,
  id: json['id'] as String?,
  employerLabel: json['employer_label'] as String?,
);

Map<String, dynamic> _$SalaryStatementCreateToJson(
  _SalaryStatementCreate instance,
) => <String, dynamic>{
  'fiscal_year': instance.fiscalYear,
  'accumulable_cents': instance.accumulableCents,
  'exempt_cents': instance.exemptCents,
  'isr_withheld_cents': instance.isrWithheldCents,
  'simultaneous': instance.simultaneous,
  'left_before_year_end': instance.leftBeforeYearEnd,
  'payer_not_withholding': instance.payerNotWithholding,
  'id': ?instance.id,
  'employer_label': ?instance.employerLabel,
};
