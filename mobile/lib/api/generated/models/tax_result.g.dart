// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'tax_result.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_TaxResult _$TaxResultFromJson(Map<String, dynamic> json) => _TaxResult(
  amountCents: (json['amount_cents'] as num).toInt(),
  payableRoundedCents: (json['payable_rounded_cents'] as num).toInt(),
  outcome: TaxResultOutcome.fromJson(json['outcome'] as String),
);

Map<String, dynamic> _$TaxResultToJson(_TaxResult instance) =>
    <String, dynamic>{
      'amount_cents': instance.amountCents,
      'payable_rounded_cents': instance.payableRoundedCents,
      'outcome': _$TaxResultOutcomeEnumMap[instance.outcome]!,
    };

const _$TaxResultOutcomeEnumMap = {
  TaxResultOutcome.payable: 'payable',
  TaxResultOutcome.inFavor: 'in_favor',
  TaxResultOutcome.zero: 'zero',
  TaxResultOutcome.$unknown: r'$unknown',
};
