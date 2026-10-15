// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'annual_tax.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_AnnualTax _$AnnualTaxFromJson(Map<String, dynamic> json) => _AnnualTax(
  accumulableIncomeCents: (json['accumulable_income_cents'] as num).toInt(),
  personalDeductionsCents: (json['personal_deductions_cents'] as num?)?.toInt(),
  baseCents: (json['base_cents'] as num).toInt(),
  isrCents: (json['isr_cents'] as num).toInt(),
  credits: AnnualCredits.fromJson(json['credits'] as Map<String, dynamic>),
  balance: AnnualBalance.fromJson(json['balance'] as Map<String, dynamic>),
);

Map<String, dynamic> _$AnnualTaxToJson(_AnnualTax instance) =>
    <String, dynamic>{
      'accumulable_income_cents': instance.accumulableIncomeCents,
      'personal_deductions_cents': instance.personalDeductionsCents,
      'base_cents': instance.baseCents,
      'isr_cents': instance.isrCents,
      'credits': instance.credits,
      'balance': instance.balance,
    };
