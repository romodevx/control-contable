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
  paidOn: json['paid_on'] as String,
  hasCfdi: json['has_cfdi'] as bool,
  cfdiUuid: json['cfdi_uuid'] as String?,
  supplierKind: json['supplier_kind'] == null
      ? null
      : SupplierKind.fromJson(json['supplier_kind'] as String),
  supplierRegime: json['supplier_regime'] == null
      ? null
      : SupplierRegime.fromJson(json['supplier_regime'] as String),
  operationKind: json['operation_kind'] == null
      ? null
      : OperationKind.fromJson(json['operation_kind'] as String),
  residentialRent: json['residential_rent'] as bool,
  isrWithheldCents: (json['isr_withheld_cents'] as num).toInt(),
  ivaWithheldCents: (json['iva_withheld_cents'] as num).toInt(),
  withholding: json['withholding'] == null
      ? null
      : Withholding.fromJson(json['withholding'] as Map<String, dynamic>),
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
  'paid_on': instance.paidOn,
  'has_cfdi': instance.hasCfdi,
  'cfdi_uuid': instance.cfdiUuid,
  'supplier_kind': _$SupplierKindEnumMap[instance.supplierKind],
  'supplier_regime': _$SupplierRegimeEnumMap[instance.supplierRegime],
  'operation_kind': _$OperationKindEnumMap[instance.operationKind],
  'residential_rent': instance.residentialRent,
  'isr_withheld_cents': instance.isrWithheldCents,
  'iva_withheld_cents': instance.ivaWithheldCents,
  'withholding': instance.withholding,
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

const _$SupplierKindEnumMap = {
  SupplierKind.pm: 'pm',
  SupplierKind.pfWithActivity: 'pf_with_activity',
  SupplierKind.foreign: 'foreign',
  SupplierKind.$unknown: r'$unknown',
};

const _$SupplierRegimeEnumMap = {
  SupplierRegime.value612: '612',
  SupplierRegime.value606: '606',
  SupplierRegime.value626: '626',
  SupplierRegime.value625: '625',
  SupplierRegime.value621: '621',
  SupplierRegime.unknown: 'unknown',
  SupplierRegime.$unknown: r'$unknown',
};

const _$OperationKindEnumMap = {
  OperationKind.professionalServices: 'professional_services',
  OperationKind.businessServices: 'business_services',
  OperationKind.rentRealEstate: 'rent_real_estate',
  OperationKind.rentMovable: 'rent_movable',
  OperationKind.saleOfGoods: 'sale_of_goods',
  OperationKind.groundFreight: 'ground_freight',
  OperationKind.commission: 'commission',
  OperationKind.scrap: 'scrap',
  OperationKind.other: 'other',
  OperationKind.$unknown: r'$unknown',
};

const _$DeductibleSourceEnumMap = {
  DeductibleSource.categoryDefault: 'category_default',
  DeductibleSource.user: 'user',
  DeductibleSource.$unknown: r'$unknown',
};
