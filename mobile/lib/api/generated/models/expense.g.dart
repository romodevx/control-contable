// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'expense.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Expense _$ExpenseFromJson(Map<String, dynamic> json) => _Expense(
  id: json['id'] as String,
  taxpayerId: json['taxpayer_id'] as String,
  regimeCode: json['regime_code'] as String,
  categoryCode: json['category_code'] as String,
  amountCents: (json['amount_cents'] as num).toInt(),
  ivaRate: IvaRate.fromJson(json['iva_rate'] as String),
  ivaCents: (json['iva_cents'] as num).toInt(),
  paymentMethod: PaymentMethod.fromJson(json['payment_method'] as String),
  paidOn: DateTime.parse(json['paid_on'] as String),
  hasCfdi: json['has_cfdi'] as bool,
  cfdiUuid: json['cfdi_uuid'] as String?,
  isrWithheldCents: (json['isr_withheld_cents'] as num).toInt(),
  ivaWithheldCents: (json['iva_withheld_cents'] as num).toInt(),
  description: json['description'] as String?,
  deductible: json['deductible'] as bool,
  deductibleSource: DeductibleSource.fromJson(
    json['deductible_source'] as String,
  ),
  warnings: (json['warnings'] as List<dynamic>)
      .map((e) => Warning.fromJson(e as Map<String, dynamic>))
      .toList(),
  legend: json['legend'] as String,
  createdAt: DateTime.parse(json['created_at'] as String),
  updatedAt: DateTime.parse(json['updated_at'] as String),
);

Map<String, dynamic> _$ExpenseToJson(_Expense instance) => <String, dynamic>{
  'id': instance.id,
  'taxpayer_id': instance.taxpayerId,
  'regime_code': instance.regimeCode,
  'category_code': instance.categoryCode,
  'amount_cents': instance.amountCents,
  'iva_rate': _$IvaRateEnumMap[instance.ivaRate]!,
  'iva_cents': instance.ivaCents,
  'payment_method': _$PaymentMethodEnumMap[instance.paymentMethod]!,
  'paid_on': instance.paidOn.toIso8601String(),
  'has_cfdi': instance.hasCfdi,
  'cfdi_uuid': instance.cfdiUuid,
  'isr_withheld_cents': instance.isrWithheldCents,
  'iva_withheld_cents': instance.ivaWithheldCents,
  'description': ?instance.description,
  'deductible': instance.deductible,
  'deductible_source': _$DeductibleSourceEnumMap[instance.deductibleSource]!,
  'warnings': instance.warnings,
  'legend': instance.legend,
  'created_at': instance.createdAt.toIso8601String(),
  'updated_at': instance.updatedAt.toIso8601String(),
};

const _$IvaRateEnumMap = {
  IvaRate.rate16: 'rate_16',
  IvaRate.rate8: 'rate_8',
  IvaRate.rate0: 'rate_0',
  IvaRate.exempt: 'exempt',
  IvaRate.$unknown: r'$unknown',
};

const _$PaymentMethodEnumMap = {
  PaymentMethod.cash: 'cash',
  PaymentMethod.transfer: 'transfer',
  PaymentMethod.cardDebit: 'card_debit',
  PaymentMethod.cardCredit: 'card_credit',
  PaymentMethod.check: 'check',
  PaymentMethod.eWallet: 'e_wallet',
  PaymentMethod.other: 'other',
  PaymentMethod.$unknown: r'$unknown',
};

const _$DeductibleSourceEnumMap = {
  DeductibleSource.categoryDefault: 'category_default',
  DeductibleSource.user: 'user',
  DeductibleSource.$unknown: r'$unknown',
};
