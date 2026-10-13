// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'deductibility_hint.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_DeductibilityHint _$DeductibilityHintFromJson(Map<String, dynamic> json) =>
    _DeductibilityHint(
      regimeCode: json['regime_code'] as String,
      deductibility: DeductibilityHintDeductibility.fromJson(
        json['deductibility'] as String,
      ),
      note: json['note'] as String?,
      legalBasis: json['legal_basis'] as String?,
    );

Map<String, dynamic> _$DeductibilityHintToJson(_DeductibilityHint instance) =>
    <String, dynamic>{
      'regime_code': instance.regimeCode,
      'deductibility':
          _$DeductibilityHintDeductibilityEnumMap[instance.deductibility]!,
      'note': ?instance.note,
      'legal_basis': ?instance.legalBasis,
    };

const _$DeductibilityHintDeductibilityEnumMap = {
  DeductibilityHintDeductibility.deductible: 'deductible',
  DeductibilityHintDeductibility.partiallyDeductible: 'partially_deductible',
  DeductibilityHintDeductibility.notDeductible: 'not_deductible',
  DeductibilityHintDeductibility.isrNotApplicable: 'isr_not_applicable',
  DeductibilityHintDeductibility.$unknown: r'$unknown',
};
