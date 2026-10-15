// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'annual_breakdown_line.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_AnnualBreakdownLine _$AnnualBreakdownLineFromJson(Map<String, dynamic> json) =>
    _AnnualBreakdownLine(
      seq: (json['seq'] as num).toInt(),
      regimeCode: json['regime_code'] as String?,
      concept: json['concept'] as String,
      operation: AnnualBreakdownLineOperation.fromJson(
        json['operation'] as String,
      ),
      amountCents: (json['amount_cents'] as num).toInt(),
      rate: json['rate'] as String?,
      legalBasis: (json['legal_basis'] as List<dynamic>)
          .map((e) => Citation.fromJson(e as Map<String, dynamic>))
          .toList(),
      source: json['source'] as String,
      needsCpaReview: json['needs_cpa_review'] as bool,
    );

Map<String, dynamic> _$AnnualBreakdownLineToJson(
  _AnnualBreakdownLine instance,
) => <String, dynamic>{
  'seq': instance.seq,
  'regime_code': instance.regimeCode,
  'concept': instance.concept,
  'operation': _$AnnualBreakdownLineOperationEnumMap[instance.operation]!,
  'amount_cents': instance.amountCents,
  'rate': instance.rate,
  'legal_basis': instance.legalBasis,
  'source': instance.source,
  'needs_cpa_review': instance.needsCpaReview,
};

const _$AnnualBreakdownLineOperationEnumMap = {
  AnnualBreakdownLineOperation.input: 'input',
  AnnualBreakdownLineOperation.add: 'add',
  AnnualBreakdownLineOperation.subtract: 'subtract',
  AnnualBreakdownLineOperation.multiply: 'multiply',
  AnnualBreakdownLineOperation.lookup: 'lookup',
  AnnualBreakdownLineOperation.result: 'result',
  AnnualBreakdownLineOperation.roundingAdjustment: 'rounding_adjustment',
  AnnualBreakdownLineOperation.info: 'info',
  AnnualBreakdownLineOperation.$unknown: r'$unknown',
};
