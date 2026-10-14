// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'trend_tax.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_TrendTax _$TrendTaxFromJson(Map<String, dynamic> json) => _TrendTax(
  ivaPayableRoundedCents: (json['iva_payable_rounded_cents'] as num).toInt(),
  isrPayableRoundedCents: (json['isr_payable_rounded_cents'] as num).toInt(),
  withholdingsToRemitCents: (json['withholdings_to_remit_cents'] as num)
      .toInt(),
  totalPayableRoundedCents: (json['total_payable_rounded_cents'] as num)
      .toInt(),
  ivaOutcome: TrendTaxIvaOutcome.fromJson(json['iva_outcome'] as String),
  status: TaxpayerEstimateStatus.fromJson(json['status'] as String),
);

Map<String, dynamic> _$TrendTaxToJson(_TrendTax instance) => <String, dynamic>{
  'iva_payable_rounded_cents': instance.ivaPayableRoundedCents,
  'isr_payable_rounded_cents': instance.isrPayableRoundedCents,
  'withholdings_to_remit_cents': instance.withholdingsToRemitCents,
  'total_payable_rounded_cents': instance.totalPayableRoundedCents,
  'iva_outcome': _$TrendTaxIvaOutcomeEnumMap[instance.ivaOutcome]!,
  'status': _$TaxpayerEstimateStatusEnumMap[instance.status]!,
};

const _$TrendTaxIvaOutcomeEnumMap = {
  TrendTaxIvaOutcome.payable: 'payable',
  TrendTaxIvaOutcome.inFavor: 'in_favor',
  TrendTaxIvaOutcome.zero: 'zero',
  TrendTaxIvaOutcome.$unknown: r'$unknown',
};

const _$TaxpayerEstimateStatusEnumMap = {
  TaxpayerEstimateStatus.complete: 'complete',
  TaxpayerEstimateStatus.incomplete: 'incomplete',
  TaxpayerEstimateStatus.partial: 'partial',
  TaxpayerEstimateStatus.informational: 'informational',
  TaxpayerEstimateStatus.none: 'none',
  TaxpayerEstimateStatus.$unknown: r'$unknown',
};
