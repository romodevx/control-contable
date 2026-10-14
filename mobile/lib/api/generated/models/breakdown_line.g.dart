// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'breakdown_line.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_BreakdownLine _$BreakdownLineFromJson(Map<String, dynamic> json) =>
    _BreakdownLine(
      seq: (json['seq'] as num).toInt(),
      tax: BreakdownLineTax.fromJson(json['tax'] as String),
      concept: json['concept'] as String,
      operation: BreakdownLineOperation.fromJson(json['operation'] as String),
      amountCents: (json['amount_cents'] as num).toInt(),
      rate: json['rate'] as String?,
      legalBasis: (json['legal_basis'] as List<dynamic>)
          .map((e) => Citation.fromJson(e as Map<String, dynamic>))
          .toList(),
      source: json['source'] as String,
      needsCpaReview: json['needs_cpa_review'] as bool,
    );

Map<String, dynamic> _$BreakdownLineToJson(_BreakdownLine instance) =>
    <String, dynamic>{
      'seq': instance.seq,
      'tax': _$BreakdownLineTaxEnumMap[instance.tax]!,
      'concept': instance.concept,
      'operation': _$BreakdownLineOperationEnumMap[instance.operation]!,
      'amount_cents': instance.amountCents,
      'rate': instance.rate,
      'legal_basis': instance.legalBasis,
      'source': instance.source,
      'needs_cpa_review': instance.needsCpaReview,
    };

const _$BreakdownLineTaxEnumMap = {
  BreakdownLineTax.isr: 'isr',
  BreakdownLineTax.iva: 'iva',
  BreakdownLineTax.withholding: 'withholding',
  BreakdownLineTax.$unknown: r'$unknown',
};

const _$BreakdownLineOperationEnumMap = {
  BreakdownLineOperation.input: 'input',
  BreakdownLineOperation.add: 'add',
  BreakdownLineOperation.subtract: 'subtract',
  BreakdownLineOperation.multiply: 'multiply',
  BreakdownLineOperation.lookup: 'lookup',
  BreakdownLineOperation.result: 'result',
  BreakdownLineOperation.roundingAdjustment: 'rounding_adjustment',
  BreakdownLineOperation.info: 'info',
  BreakdownLineOperation.$unknown: r'$unknown',
};
