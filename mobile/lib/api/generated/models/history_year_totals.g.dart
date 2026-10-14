// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'history_year_totals.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_HistoryYearTotals _$HistoryYearTotalsFromJson(
  Map<String, dynamic> json,
) => _HistoryYearTotals(
  isrPayableRoundedCents: (json['isr_payable_rounded_cents'] as num).toInt(),
  ivaPayableRoundedCents: (json['iva_payable_rounded_cents'] as num).toInt(),
  withholdingsToRemitCents: (json['withholdings_to_remit_cents'] as num)
      .toInt(),
  totalPayableRoundedCents: (json['total_payable_rounded_cents'] as num)
      .toInt(),
);

Map<String, dynamic> _$HistoryYearTotalsToJson(_HistoryYearTotals instance) =>
    <String, dynamic>{
      'isr_payable_rounded_cents': instance.isrPayableRoundedCents,
      'iva_payable_rounded_cents': instance.ivaPayableRoundedCents,
      'withholdings_to_remit_cents': instance.withholdingsToRemitCents,
      'total_payable_rounded_cents': instance.totalPayableRoundedCents,
    };
