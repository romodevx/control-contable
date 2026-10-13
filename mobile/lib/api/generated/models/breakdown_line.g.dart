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
      legalBasis: (json['legal_basis'] as List<dynamic>)
          .map((e) => Citation.fromJson(e as Map<String, dynamic>))
          .toList(),
      rate: json['rate'] as String?,
    );

Map<String, dynamic> _$BreakdownLineToJson(_BreakdownLine instance) =>
    <String, dynamic>{
      'seq': instance.seq,
      'tax': _$BreakdownLineTaxEnumMap[instance.tax]!,
      'concept': instance.concept,
      'operation': _$BreakdownLineOperationEnumMap[instance.operation]!,
      'amount_cents': instance.amountCents,
      'legal_basis': instance.legalBasis,
      'rate': instance.rate,
    };

const _$BreakdownLineTaxEnumMap = {
  BreakdownLineTax.iva: 'iva',
  BreakdownLineTax.isr: 'isr',
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
