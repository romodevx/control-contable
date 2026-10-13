// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'disclaimer.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Disclaimer _$DisclaimerFromJson(Map<String, dynamic> json) => _Disclaimer(
  code: DisclaimerCode.fromJson(json['code'] as String),
  text: json['text'] as String,
);

Map<String, dynamic> _$DisclaimerToJson(_Disclaimer instance) =>
    <String, dynamic>{
      'code': _$DisclaimerCodeEnumMap[instance.code]!,
      'text': instance.text,
    };

const _$DisclaimerCodeEnumMap = {
  DisclaimerCode.notTaxAdvice: 'not_tax_advice',
  DisclaimerCode.incompleteInputs: 'incomplete_inputs',
  DisclaimerCode.provisionalParameters: 'provisional_parameters',
  DisclaimerCode.informationalRegime: 'informational_regime',
  DisclaimerCode.deductibilityClaimed: 'deductibility_claimed',
  DisclaimerCode.noCoefficientFirstYear: 'no_coefficient_first_year',
  DisclaimerCode.$unknown: r'$unknown',
};
