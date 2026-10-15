// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'salary_statement.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_SalaryStatement _$SalaryStatementFromJson(Map<String, dynamic> json) =>
    _SalaryStatement(
      id: json['id'] as String,
      taxpayerId: json['taxpayer_id'] as String,
      fiscalYear: (json['fiscal_year'] as num).toInt(),
      employerLabel: json['employer_label'] as String?,
      accumulableCents: (json['accumulable_cents'] as num).toInt(),
      exemptCents: (json['exempt_cents'] as num).toInt(),
      isrWithheldCents: (json['isr_withheld_cents'] as num).toInt(),
      simultaneous: json['simultaneous'] as bool,
      leftBeforeYearEnd: json['left_before_year_end'] as bool,
      payerNotWithholding: json['payer_not_withholding'] as bool,
      warnings: (json['warnings'] as List<dynamic>)
          .map((e) => Warning.fromJson(e as Map<String, dynamic>))
          .toList(),
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: DateTime.parse(json['updated_at'] as String),
    );

Map<String, dynamic> _$SalaryStatementToJson(_SalaryStatement instance) =>
    <String, dynamic>{
      'id': instance.id,
      'taxpayer_id': instance.taxpayerId,
      'fiscal_year': instance.fiscalYear,
      'employer_label': instance.employerLabel,
      'accumulable_cents': instance.accumulableCents,
      'exempt_cents': instance.exemptCents,
      'isr_withheld_cents': instance.isrWithheldCents,
      'simultaneous': instance.simultaneous,
      'left_before_year_end': instance.leftBeforeYearEnd,
      'payer_not_withholding': instance.payerNotWithholding,
      'warnings': instance.warnings,
      'created_at': instance.createdAt.toIso8601String(),
      'updated_at': instance.updatedAt.toIso8601String(),
    };
