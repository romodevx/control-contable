// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'resico_pf_check.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_ResicoPfCheck _$ResicoPfCheckFromJson(Map<String, dynamic> json) =>
    _ResicoPfCheck(
      incomeCents: (json['income_cents'] as num).toInt(),
      rate: json['rate'] as String,
      annualIsrCents: (json['annual_isr_cents'] as num).toInt(),
      monthlyPaymentsCents: (json['monthly_payments_cents'] as num).toInt(),
      withheldCents: (json['withheld_cents'] as num).toInt(),
      differenceCents: (json['difference_cents'] as num).toInt(),
    );

Map<String, dynamic> _$ResicoPfCheckToJson(_ResicoPfCheck instance) =>
    <String, dynamic>{
      'income_cents': instance.incomeCents,
      'rate': instance.rate,
      'annual_isr_cents': instance.annualIsrCents,
      'monthly_payments_cents': instance.monthlyPaymentsCents,
      'withheld_cents': instance.withheldCents,
      'difference_cents': instance.differenceCents,
    };
