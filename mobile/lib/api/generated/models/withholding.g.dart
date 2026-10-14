// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'withholding.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Withholding _$WithholdingFromJson(Map<String, dynamic> json) => _Withholding(
  isrCents: (json['isr_cents'] as num).toInt(),
  ivaCents: (json['iva_cents'] as num).toInt(),
  isrRate: json['isr_rate'] as String,
  ivaBasis: WithholdingBasis.fromJson(json['iva_basis'] as String),
  withholder: Withholder.fromJson(json['withholder'] as String),
  netCents: (json['net_cents'] as num).toInt(),
  remitDueOn: json['remit_due_on'] as String?,
  ruleIds: (json['rule_ids'] as List<dynamic>).map((e) => e as String).toList(),
  explanations: (json['explanations'] as List<dynamic>)
      .map((e) => WithholdingExplanation.fromJson(e as Map<String, dynamic>))
      .toList(),
  tableVersion: json['table_version'] as String,
  cpaUnverified: json['cpa_unverified'] as bool,
);

Map<String, dynamic> _$WithholdingToJson(_Withholding instance) =>
    <String, dynamic>{
      'isr_cents': instance.isrCents,
      'iva_cents': instance.ivaCents,
      'isr_rate': instance.isrRate,
      'iva_basis': _$WithholdingBasisEnumMap[instance.ivaBasis]!,
      'withholder': _$WithholderEnumMap[instance.withholder]!,
      'net_cents': instance.netCents,
      'remit_due_on': instance.remitDueOn,
      'rule_ids': instance.ruleIds,
      'explanations': instance.explanations,
      'table_version': instance.tableVersion,
      'cpa_unverified': instance.cpaUnverified,
    };

const _$WithholdingBasisEnumMap = {
  WithholdingBasis.none: 'none',
  WithholdingBasis.twoThirdsOfIva: 'two_thirds_of_iva',
  WithholdingBasis.halfOfIva: 'half_of_iva',
  WithholdingBasis.allOfIva: 'all_of_iva',
  WithholdingBasis.fourPercentOfPrice: 'four_percent_of_price',
  WithholdingBasis.$unknown: r'$unknown',
};

const _$WithholderEnumMap = {
  Withholder.counterparty: 'counterparty',
  Withholder.you: 'you',
  Withholder.platform: 'platform',
  Withholder.none: 'none',
  Withholder.$unknown: r'$unknown',
};
