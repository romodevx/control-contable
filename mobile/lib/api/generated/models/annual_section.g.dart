// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'annual_section.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_AnnualSection _$AnnualSectionFromJson(
  Map<String, dynamic> json,
) => _AnnualSection(
  kind: AnnualSectionKind.fromJson(json['kind'] as String),
  regimes: (json['regimes'] as List<dynamic>).map((e) => e as String).toList(),
  status: AnnualSectionStatus.fromJson(json['status'] as String),
  missingInputs: (json['missing_inputs'] as List<dynamic>)
      .map((e) => AnnualMissingInput.fromJson(e as Map<String, dynamic>))
      .toList(),
  obligation: json['obligation'] == null
      ? null
      : AnnualObligation.fromJson(json['obligation'] as String),
  obligationReasons: (json['obligation_reasons'] as List<dynamic>)
      .map((e) => e as String)
      .toList(),
  deadline: json['deadline'] as String?,
  tax: json['tax'] == null
      ? null
      : AnnualTax.fromJson(json['tax'] as Map<String, dynamic>),
  pmExtras: json['pm_extras'] == null
      ? null
      : PmAnnualExtras.fromJson(json['pm_extras'] as Map<String, dynamic>),
  resicoCheck: json['resico_check'] == null
      ? null
      : ResicoPfCheck.fromJson(json['resico_check'] as Map<String, dynamic>),
  explanation: json['explanation'] as String?,
  notes: (json['notes'] as List<dynamic>)
      .map((e) => AnnualNote.fromJson(e as Map<String, dynamic>))
      .toList(),
  cpaReviewPending: json['cpa_review_pending'] as bool,
  cpaTopics: (json['cpa_topics'] as List<dynamic>)
      .map((e) => CpaTopic.fromJson(e as String))
      .toList(),
  parameterSet: ParameterSetRef.fromJson(
    json['parameter_set'] as Map<String, dynamic>,
  ),
);

Map<String, dynamic> _$AnnualSectionToJson(
  _AnnualSection instance,
) => <String, dynamic>{
  'kind': _$AnnualSectionKindEnumMap[instance.kind]!,
  'regimes': instance.regimes,
  'status': _$AnnualSectionStatusEnumMap[instance.status]!,
  'missing_inputs': instance.missingInputs,
  'obligation': _$AnnualObligationEnumMap[instance.obligation],
  'obligation_reasons': instance.obligationReasons,
  'deadline': instance.deadline,
  'tax': instance.tax,
  'pm_extras': instance.pmExtras,
  'resico_check': instance.resicoCheck,
  'explanation': instance.explanation,
  'notes': instance.notes,
  'cpa_review_pending': instance.cpaReviewPending,
  'cpa_topics': instance.cpaTopics.map((e) => _$CpaTopicEnumMap[e]!).toList(),
  'parameter_set': instance.parameterSet,
};

const _$AnnualSectionKindEnumMap = {
  AnnualSectionKind.pfArt152: 'pf_art152',
  AnnualSectionKind.pfResicoCheck: 'pf_resico_check',
  AnnualSectionKind.pm601: 'pm_601',
  AnnualSectionKind.pmResico: 'pm_resico',
  AnnualSectionKind.explanation: 'explanation',
  AnnualSectionKind.$unknown: r'$unknown',
};

const _$AnnualSectionStatusEnumMap = {
  AnnualSectionStatus.complete: 'complete',
  AnnualSectionStatus.incomplete: 'incomplete',
  AnnualSectionStatus.informational: 'informational',
  AnnualSectionStatus.$unknown: r'$unknown',
};

const _$AnnualObligationEnumMap = {
  AnnualObligation.obligated: 'obligated',
  AnnualObligation.optional: 'optional',
  AnnualObligation.notObligated: 'not_obligated',
  AnnualObligation.relieved: 'relieved',
  AnnualObligation.$unknown: r'$unknown',
};

const _$CpaTopicEnumMap = {
  CpaTopic.rifReduction: 'rif_reduction',
  CpaTopic.rifBimonthlyIva: 'rif_bimonthly_iva',
  CpaTopic.platformDefinitiveIva: 'platform_definitive_iva',
  CpaTopic.crossRegimeIvaNetting: 'cross_regime_iva_netting',
  CpaTopic.resicoPmInterestWithholding: 'resico_pm_interest_withholding',
  CpaTopic.withholdingRuleUnverified: 'withholding_rule_unverified',
  CpaTopic.withheldIvaCreditTiming: 'withheld_iva_credit_timing',
  CpaTopic.platformNondefinitiveAnnual: 'platform_nondefinitive_annual',
  CpaTopic.umaForMinimumWage: 'uma_for_minimum_wage',
  CpaTopic.deductionCapResicoIncome: 'deduction_cap_resico_income',
  CpaTopic.salaryObligationBasis: 'salary_obligation_basis',
  CpaTopic.cuRounding: 'cu_rounding',
  CpaTopic.resicoPmPtuBase: 'resico_pm_ptu_base',
  CpaTopic.resicoPfReliefRenewal: 'resico_pf_relief_renewal',
  CpaTopic.ptuRate: 'ptu_rate',
  CpaTopic.$unknown: r'$unknown',
};
