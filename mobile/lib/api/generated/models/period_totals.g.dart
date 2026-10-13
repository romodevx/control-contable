// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'period_totals.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_PeriodTotals _$PeriodTotalsFromJson(Map<String, dynamic> json) =>
    _PeriodTotals(
      count: (json['count'] as num).toInt(),
      amountCents: (json['amount_cents'] as num).toInt(),
      ivaCents: (json['iva_cents'] as num).toInt(),
      isrWithheldCents: (json['isr_withheld_cents'] as num?)?.toInt(),
      ivaWithheldCents: (json['iva_withheld_cents'] as num?)?.toInt(),
    );

Map<String, dynamic> _$PeriodTotalsToJson(_PeriodTotals instance) =>
    <String, dynamic>{
      'count': instance.count,
      'amount_cents': instance.amountCents,
      'iva_cents': instance.ivaCents,
      'isr_withheld_cents': instance.isrWithheldCents,
      'iva_withheld_cents': instance.ivaWithheldCents,
    };
