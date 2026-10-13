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
      parameterSet: ParameterSetRef.fromJson(
        json['parameter_set'] as Map<String, dynamic>,
      ),
      missingInputs: (json['missing_inputs'] as List<dynamic>?)
          ?.map((e) => e as String)
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
      withheldToRemitCents: (json['withheld_to_remit_cents'] as num?)?.toInt(),
    );

Map<String, dynamic> _$RegimeEstimateToJson(_RegimeEstimate instance) =>
    <String, dynamic>{
      'regime_code': instance.regimeCode,
      'regime_name': instance.regimeName,
      'treatment': _$TreatmentEnumMap[instance.treatment]!,
      'status': _$RegimeEstimateStatusEnumMap[instance.status]!,
      'parameter_set': instance.parameterSet,
      'missing_inputs': ?instance.missingInputs,
      'explanation': ?instance.explanation,
      'isr': ?instance.isr,
      'iva_components': ?instance.ivaComponents,
      'withheld_to_remit_cents': ?instance.withheldToRemitCents,
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
  RegimeEstimateStatus.informational: 'informational',
  RegimeEstimateStatus.$unknown: r'$unknown',
};
