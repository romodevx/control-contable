// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'warning.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Warning _$WarningFromJson(Map<String, dynamic> json) => _Warning(
  code: WarningCode.fromJson(json['code'] as String),
  field: json['field'] as String,
  message: json['message'] as String,
);

Map<String, dynamic> _$WarningToJson(_Warning instance) => <String, dynamic>{
  'code': _$WarningCodeEnumMap[instance.code]!,
  'field': instance.field,
  'message': instance.message,
};

const _$WarningCodeEnumMap = {
  WarningCode.rfcCheckDigitMismatch: 'rfc_check_digit_mismatch',
  WarningCode.regimeCombinationNeedsCpaReview:
      'regime_combination_needs_cpa_review',
  WarningCode.duplicateCfdiUuid: 'duplicate_cfdi_uuid',
  WarningCode.withholdingDiffersFromRate: 'withholding_differs_from_rate',
  WarningCode.cashOverThreshold: 'cash_over_threshold',
  WarningCode.noCfdi: 'no_cfdi',
  WarningCode.notIsrDeductibleInRegime: 'not_isr_deductible_in_regime',
  WarningCode.categoryNotDeductible: 'category_not_deductible',
  WarningCode.ivaNotCreditableCashPayment: 'iva_not_creditable_cash_payment',
  WarningCode.$unknown: r'$unknown',
};
