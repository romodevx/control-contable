// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'income.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Income _$IncomeFromJson(Map<String, dynamic> json) => _Income(
  id: json['id'] as String,
  taxpayerId: json['taxpayer_id'] as String,
  regimeCode: json['regime_code'] as String,
  counterpartyKind: CounterpartyKind.fromJson(
    json['counterparty_kind'] as String,
  ),
  operationKind: OperationKind.fromJson(json['operation_kind'] as String),
  residentialRent: json['residential_rent'] as bool,
  viaPlatform: json['via_platform'] as bool,
  amountCents: (json['amount_cents'] as num).toInt(),
  ivaRate: IvaRate.fromJson(json['iva_rate'] as String),
  ivaCents: (json['iva_cents'] as num).toInt(),
  isrWithheldCents: (json['isr_withheld_cents'] as num).toInt(),
  ivaWithheldCents: (json['iva_withheld_cents'] as num).toInt(),
  withholding: Withholding.fromJson(
    json['withholding'] as Map<String, dynamic>,
  ),
  invoicedOn: json['invoiced_on'] as String?,
  collectedOn: json['collected_on'] as String?,
  platformActivity: json['platform_activity'] == null
      ? null
      : PlatformActivity.fromJson(json['platform_activity'] as String),
  cfdiUuid: json['cfdi_uuid'] as String?,
  description: json['description'] as String?,
  pendingCollection: json['pending_collection'] as bool,
  warnings: (json['warnings'] as List<dynamic>)
      .map((e) => Warning.fromJson(e as Map<String, dynamic>))
      .toList(),
  createdAt: DateTime.parse(json['created_at'] as String),
  updatedAt: DateTime.parse(json['updated_at'] as String),
);

Map<String, dynamic> _$IncomeToJson(_Income instance) => <String, dynamic>{
  'id': instance.id,
  'taxpayer_id': instance.taxpayerId,
  'regime_code': instance.regimeCode,
  'counterparty_kind': _$CounterpartyKindEnumMap[instance.counterpartyKind]!,
  'operation_kind': _$OperationKindEnumMap[instance.operationKind]!,
  'residential_rent': instance.residentialRent,
  'via_platform': instance.viaPlatform,
  'amount_cents': instance.amountCents,
  'iva_rate': _$IvaRateEnumMap[instance.ivaRate]!,
  'iva_cents': instance.ivaCents,
  'isr_withheld_cents': instance.isrWithheldCents,
  'iva_withheld_cents': instance.ivaWithheldCents,
  'withholding': instance.withholding,
  'invoiced_on': instance.invoicedOn,
  'collected_on': instance.collectedOn,
  'platform_activity': _$PlatformActivityEnumMap[instance.platformActivity],
  'cfdi_uuid': instance.cfdiUuid,
  'description': ?instance.description,
  'pending_collection': instance.pendingCollection,
  'warnings': instance.warnings,
  'created_at': instance.createdAt.toIso8601String(),
  'updated_at': instance.updatedAt.toIso8601String(),
};

const _$CounterpartyKindEnumMap = {
  CounterpartyKind.pm: 'pm',
  CounterpartyKind.pfWithActivity: 'pf_with_activity',
  CounterpartyKind.pfNoActivityOrPublic: 'pf_no_activity_or_public',
  CounterpartyKind.govFederal: 'gov_federal',
  CounterpartyKind.govStateMunicipal: 'gov_state_municipal',
  CounterpartyKind.foreign: 'foreign',
  CounterpartyKind.$unknown: r'$unknown',
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

const _$IvaRateEnumMap = {
  IvaRate.rate16: 'rate_16',
  IvaRate.rate8: 'rate_8',
  IvaRate.rate0: 'rate_0',
  IvaRate.exempt: 'exempt',
  IvaRate.$unknown: r'$unknown',
};

const _$PlatformActivityEnumMap = {
  PlatformActivity.transportDelivery: 'transport_delivery',
  PlatformActivity.lodging: 'lodging',
  PlatformActivity.goodsServices: 'goods_services',
  PlatformActivity.$unknown: r'$unknown',
};
