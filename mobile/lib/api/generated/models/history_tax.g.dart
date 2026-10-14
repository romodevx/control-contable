// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'history_tax.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_HistoryTax _$HistoryTaxFromJson(Map<String, dynamic> json) => _HistoryTax(
  isrPayableRoundedCents: (json['isr_payable_rounded_cents'] as num).toInt(),
  ivaPayableRoundedCents: (json['iva_payable_rounded_cents'] as num).toInt(),
  ivaOutcome: HistoryTaxIvaOutcome.fromJson(json['iva_outcome'] as String),
  ivaInFavorCents: (json['iva_in_favor_cents'] as num).toInt(),
  withholdingsToRemitCents: (json['withholdings_to_remit_cents'] as num)
      .toInt(),
  totalPayableRoundedCents: (json['total_payable_rounded_cents'] as num)
      .toInt(),
);

Map<String, dynamic> _$HistoryTaxToJson(_HistoryTax instance) =>
    <String, dynamic>{
      'isr_payable_rounded_cents': instance.isrPayableRoundedCents,
      'iva_payable_rounded_cents': instance.ivaPayableRoundedCents,
      'iva_outcome': _$HistoryTaxIvaOutcomeEnumMap[instance.ivaOutcome]!,
      'iva_in_favor_cents': instance.ivaInFavorCents,
      'withholdings_to_remit_cents': instance.withholdingsToRemitCents,
      'total_payable_rounded_cents': instance.totalPayableRoundedCents,
    };

const _$HistoryTaxIvaOutcomeEnumMap = {
  HistoryTaxIvaOutcome.payable: 'payable',
  HistoryTaxIvaOutcome.inFavor: 'in_favor',
  HistoryTaxIvaOutcome.zero: 'zero',
  HistoryTaxIvaOutcome.$unknown: r'$unknown',
};
