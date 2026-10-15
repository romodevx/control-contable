// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'annual_credits.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_AnnualCredits _$AnnualCreditsFromJson(Map<String, dynamic> json) =>
    _AnnualCredits(
      paymentsCents: (json['payments_cents'] as num).toInt(),
      withholdingsCents: (json['withholdings_cents'] as num).toInt(),
      salaryIsrCents: (json['salary_isr_cents'] as num).toInt(),
      totalCents: (json['total_cents'] as num).toInt(),
      paymentSources: PaymentSourceCounts.fromJson(
        json['payment_sources'] as Map<String, dynamic>,
      ),
    );

Map<String, dynamic> _$AnnualCreditsToJson(_AnnualCredits instance) =>
    <String, dynamic>{
      'payments_cents': instance.paymentsCents,
      'withholdings_cents': instance.withholdingsCents,
      'salary_isr_cents': instance.salaryIsrCents,
      'total_cents': instance.totalCents,
      'payment_sources': instance.paymentSources,
    };
