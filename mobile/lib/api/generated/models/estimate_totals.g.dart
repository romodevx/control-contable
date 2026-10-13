// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'estimate_totals.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_EstimateTotals _$EstimateTotalsFromJson(Map<String, dynamic> json) =>
    _EstimateTotals(
      iva: TaxResult.fromJson(json['iva'] as Map<String, dynamic>),
      isr: TaxResult.fromJson(json['isr'] as Map<String, dynamic>),
      withholdingsToRemitCents: (json['withholdings_to_remit_cents'] as num)
          .toInt(),
      totalPayableRoundedCents: (json['total_payable_rounded_cents'] as num)
          .toInt(),
    );

Map<String, dynamic> _$EstimateTotalsToJson(_EstimateTotals instance) =>
    <String, dynamic>{
      'iva': instance.iva,
      'isr': instance.isr,
      'withholdings_to_remit_cents': instance.withholdingsToRemitCents,
      'total_payable_rounded_cents': instance.totalPayableRoundedCents,
    };
