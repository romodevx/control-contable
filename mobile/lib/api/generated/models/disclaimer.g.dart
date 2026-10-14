// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'disclaimer.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Disclaimer _$DisclaimerFromJson(Map<String, dynamic> json) => _Disclaimer(
  code: DisclaimerCode.fromJson(json['code'] as String),
  severity: DisclaimerSeverity.fromJson(json['severity'] as String),
  text: json['text'] as String,
);

Map<String, dynamic> _$DisclaimerToJson(_Disclaimer instance) =>
    <String, dynamic>{
      'code': _$DisclaimerCodeEnumMap[instance.code]!,
      'severity': _$DisclaimerSeverityEnumMap[instance.severity]!,
      'text': instance.text,
    };

const _$DisclaimerCodeEnumMap = {
  DisclaimerCode.notTaxAdvice: 'not_tax_advice',
  DisclaimerCode.incompleteInputs: 'incomplete_inputs',
  DisclaimerCode.provisionalParameters: 'provisional_parameters',
  DisclaimerCode.informationalRegime: 'informational_regime',
  DisclaimerCode.partialRegime: 'partial_regime',
  DisclaimerCode.deductibilityClaimed: 'deductibility_claimed',
  DisclaimerCode.noCoefficientFirstYear: 'no_coefficient_first_year',
  DisclaimerCode.ivaMixedActivity: 'iva_mixed_activity',
  DisclaimerCode.cpaReviewPending: 'cpa_review_pending',
  DisclaimerCode.$unknown: r'$unknown',
};

const _$DisclaimerSeverityEnumMap = {
  DisclaimerSeverity.info: 'info',
  DisclaimerSeverity.warning: 'warning',
  DisclaimerSeverity.strong: 'strong',
  DisclaimerSeverity.$unknown: r'$unknown',
};
