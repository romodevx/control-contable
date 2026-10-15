// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'deduction_category_total.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_DeductionCategoryTotal _$DeductionCategoryTotalFromJson(
  Map<String, dynamic> json,
) => _DeductionCategoryTotal(
  category: DeductionCategory.fromJson(json['category'] as String),
  registeredCents: (json['registered_cents'] as num).toInt(),
  excludedCents: (json['excluded_cents'] as num).toInt(),
  count: (json['count'] as num).toInt(),
);

Map<String, dynamic> _$DeductionCategoryTotalToJson(
  _DeductionCategoryTotal instance,
) => <String, dynamic>{
  'category': _$DeductionCategoryEnumMap[instance.category]!,
  'registered_cents': instance.registeredCents,
  'excluded_cents': instance.excludedCents,
  'count': instance.count,
};

const _$DeductionCategoryEnumMap = {
  DeductionCategory.medical: 'medical',
  DeductionCategory.medicalDisability: 'medical_disability',
  DeductionCategory.funeral: 'funeral',
  DeductionCategory.donation: 'donation',
  DeductionCategory.mortgageRealInterest: 'mortgage_real_interest',
  DeductionCategory.retirementContribution: 'retirement_contribution',
  DeductionCategory.healthInsurance: 'health_insurance',
  DeductionCategory.schoolTransport: 'school_transport',
  DeductionCategory.localPayrollTax: 'local_payroll_tax',
  DeductionCategory.tuition: 'tuition',
  DeductionCategory.$unknown: r'$unknown',
};
