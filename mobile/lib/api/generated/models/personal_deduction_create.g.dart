// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'personal_deduction_create.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_PersonalDeductionCreate _$PersonalDeductionCreateFromJson(
  Map<String, dynamic> json,
) => _PersonalDeductionCreate(
  category: DeductionCategory.fromJson(json['category'] as String),
  amountCents: (json['amount_cents'] as num).toInt(),
  paidOn: json['paid_on'] as String,
  paymentMethod: DeductionPaymentMethod.fromJson(
    json['payment_method'] as String,
  ),
  beneficiary: DeductionBeneficiary.fromJson(json['beneficiary'] as String),
  id: json['id'] as String?,
  doneeType: json['donee_type'] == null
      ? null
      : DoneeType.fromJson(json['donee_type'] as String),
  studentLabel: json['student_label'] as String?,
  tuitionLevel: json['tuition_level'] == null
      ? null
      : TuitionLevel.fromJson(json['tuition_level'] as String),
  scholarshipCents: (json['scholarship_cents'] as num?)?.toInt(),
  cfdiUuid: json['cfdi_uuid'] as String?,
);

Map<String, dynamic> _$PersonalDeductionCreateToJson(
  _PersonalDeductionCreate instance,
) => <String, dynamic>{
  'category': _$DeductionCategoryEnumMap[instance.category]!,
  'amount_cents': instance.amountCents,
  'paid_on': instance.paidOn,
  'payment_method': _$DeductionPaymentMethodEnumMap[instance.paymentMethod]!,
  'beneficiary': _$DeductionBeneficiaryEnumMap[instance.beneficiary]!,
  'id': ?instance.id,
  'donee_type': ?_$DoneeTypeEnumMap[instance.doneeType],
  'student_label': ?instance.studentLabel,
  'tuition_level': ?_$TuitionLevelEnumMap[instance.tuitionLevel],
  'scholarship_cents': ?instance.scholarshipCents,
  'cfdi_uuid': ?instance.cfdiUuid,
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

const _$DeductionPaymentMethodEnumMap = {
  DeductionPaymentMethod.electronic: 'electronic',
  DeductionPaymentMethod.cash: 'cash',
  DeductionPaymentMethod.$unknown: r'$unknown',
};

const _$DeductionBeneficiaryEnumMap = {
  DeductionBeneficiary.self: 'self',
  DeductionBeneficiary.spouse: 'spouse',
  DeductionBeneficiary.ascendant: 'ascendant',
  DeductionBeneficiary.descendant: 'descendant',
  DeductionBeneficiary.$unknown: r'$unknown',
};

const _$DoneeTypeEnumMap = {
  DoneeType.authorized: 'authorized',
  DoneeType.government: 'government',
  DoneeType.$unknown: r'$unknown',
};

const _$TuitionLevelEnumMap = {
  TuitionLevel.preescolar: 'preescolar',
  TuitionLevel.primaria: 'primaria',
  TuitionLevel.secundaria: 'secundaria',
  TuitionLevel.profesionalTecnico: 'profesional_tecnico',
  TuitionLevel.bachillerato: 'bachillerato',
  TuitionLevel.$unknown: r'$unknown',
};
