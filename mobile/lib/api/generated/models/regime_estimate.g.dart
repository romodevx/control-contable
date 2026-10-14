// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'regime_estimate.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_RegimeEstimate _$RegimeEstimateFromJson(Map<String, dynamic> json) =>
    _RegimeEstimate(
      regimeCode: json['regime_code'] as String,
      regimeName: json['regime_name'] as String,
      treatment: Treatment.fromJson(json['treatment'] as String),
      status: RegimeEstimateStatus.fromJson(json['status'] as String),
      missingInputs: (json['missing_inputs'] as List<dynamic>)
          .map((e) => MissingInput.fromJson(e as Map<String, dynamic>))
          .toList(),
      explanation: json['explanation'] as String?,
      isr: json['isr'] == null
          ? null
          : TaxResult.fromJson(json['isr'] as Map<String, dynamic>),
      ivaComponents: json['iva_components'] == null
          ? null
          : IvaComponents.fromJson(
              json['iva_components'] as Map<String, dynamic>,
            ),
      withheldToRemitCents: (json['withheld_to_remit_cents'] as num).toInt(),
      cpaReviewPending: json['cpa_review_pending'] as bool,
      cpaTopics: (json['cpa_topics'] as List<dynamic>)
          .map((e) => CpaTopic.fromJson(e as String))
          .toList(),
      parameterSet: ParameterSetRef.fromJson(
        json['parameter_set'] as Map<String, dynamic>,
      ),
    );

Map<String, dynamic> _$RegimeEstimateToJson(
  _RegimeEstimate instance,
) => <String, dynamic>{
  'regime_code': instance.regimeCode,
  'regime_name': instance.regimeName,
  'treatment': _$TreatmentEnumMap[instance.treatment]!,
  'status': _$RegimeEstimateStatusEnumMap[instance.status]!,
  'missing_inputs': instance.missingInputs,
  'explanation': instance.explanation,
  'isr': instance.isr,
  'iva_components': instance.ivaComponents,
  'withheld_to_remit_cents': instance.withheldToRemitCents,
  'cpa_review_pending': instance.cpaReviewPending,
  'cpa_topics': instance.cpaTopics.map((e) => _$CpaTopicEnumMap[e]!).toList(),
  'parameter_set': instance.parameterSet,
};

const _$TreatmentEnumMap = {
  Treatment.full: 'full',
  Treatment.partial: 'partial',
  Treatment.informational: 'informational',
  Treatment.$unknown: r'$unknown',
};

const _$RegimeEstimateStatusEnumMap = {
  RegimeEstimateStatus.complete: 'complete',
  RegimeEstimateStatus.incomplete: 'incomplete',
  RegimeEstimateStatus.partial: 'partial',
  RegimeEstimateStatus.informational: 'informational',
  RegimeEstimateStatus.$unknown: r'$unknown',
};

const _$CpaTopicEnumMap = {
  CpaTopic.rifReduction: 'rif_reduction',
  CpaTopic.rifBimonthlyIva: 'rif_bimonthly_iva',
  CpaTopic.platformDefinitiveIva: 'platform_definitive_iva',
  CpaTopic.crossRegimeIvaNetting: 'cross_regime_iva_netting',
  CpaTopic.resicoPmInterestWithholding: 'resico_pm_interest_withholding',
  CpaTopic.withholdingRuleUnverified: 'withholding_rule_unverified',
  CpaTopic.withheldIvaCreditTiming: 'withheld_iva_credit_timing',
  CpaTopic.$unknown: r'$unknown',
};
