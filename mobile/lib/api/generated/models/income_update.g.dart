// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'income_update.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_IncomeUpdate _$IncomeUpdateFromJson(Map<String, dynamic> json) =>
    _IncomeUpdate(
      regimeCode: json['regime_code'] as String?,
      counterpartyKind: json['counterparty_kind'] == null
          ? null
          : CounterpartyKind.fromJson(json['counterparty_kind'] as String),
      operationKind: json['operation_kind'] == null
          ? null
          : OperationKind.fromJson(json['operation_kind'] as String),
      residentialRent: json['residential_rent'] as bool?,
      viaPlatform: json['via_platform'] as bool?,
      amountCents: (json['amount_cents'] as num?)?.toInt(),
      ivaRate: json['iva_rate'] == null
          ? null
          : IvaRate.fromJson(json['iva_rate'] as String),
      ivaCents: (json['iva_cents'] as num?)?.toInt(),
      invoicedOn: json['invoiced_on'] as String?,
      collectedOn: json['collected_on'] as String?,
      platformActivity: json['platform_activity'] == null
          ? null
          : PlatformActivity.fromJson(json['platform_activity'] as String),
      cfdiUuid: json['cfdi_uuid'] as String?,
      description: json['description'] as String?,
    );

Map<String, dynamic> _$IncomeUpdateToJson(
  _IncomeUpdate instance,
) => <String, dynamic>{
  'regime_code': ?instance.regimeCode,
  'counterparty_kind': ?_$CounterpartyKindEnumMap[instance.counterpartyKind],
  'operation_kind': ?_$OperationKindEnumMap[instance.operationKind],
  'residential_rent': ?instance.residentialRent,
  'via_platform': ?instance.viaPlatform,
  'amount_cents': ?instance.amountCents,
  'iva_rate': ?_$IvaRateEnumMap[instance.ivaRate],
  'iva_cents': ?instance.ivaCents,
  'invoiced_on': ?instance.invoicedOn,
  'collected_on': ?instance.collectedOn,
  'platform_activity': ?_$PlatformActivityEnumMap[instance.platformActivity],
  'cfdi_uuid': ?instance.cfdiUuid,
  'description': ?instance.description,
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
