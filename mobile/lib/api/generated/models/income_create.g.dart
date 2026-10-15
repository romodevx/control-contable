// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'income_create.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_IncomeCreate _$IncomeCreateFromJson(Map<String, dynamic> json) =>
    _IncomeCreate(
      regimeCode: json['regime_code'] as String,
      counterpartyType: CounterpartyType.fromJson(
        json['counterparty_type'] as String,
      ),
      amountCents: (json['amount_cents'] as num).toInt(),
      ivaRate: IvaRate.fromJson(json['iva_rate'] as String),
      ivaCents: (json['iva_cents'] as num).toInt(),
      id: json['id'] as String?,
      counterpartyKind: json['counterparty_kind'] == null
          ? null
          : CounterpartyKind.fromJson(json['counterparty_kind'] as String),
      operationKind: json['operation_kind'] == null
          ? null
          : OperationKind.fromJson(json['operation_kind'] as String),
      residentialRent: json['residential_rent'] as bool?,
      viaPlatform: json['via_platform'] as bool?,
      isrWithheldCents: (json['isr_withheld_cents'] as num?)?.toInt(),
      ivaWithheldCents: (json['iva_withheld_cents'] as num?)?.toInt(),
      invoicedOn: json['invoiced_on'] as String?,
      collectedOn: json['collected_on'] as String?,
      platformActivity: json['platform_activity'] == null
          ? null
          : PlatformActivity.fromJson(json['platform_activity'] as String),
      cfdiUuid: json['cfdi_uuid'] as String?,
      description: json['description'] as String?,
    );

Map<String, dynamic> _$IncomeCreateToJson(
  _IncomeCreate instance,
) => <String, dynamic>{
  'regime_code': instance.regimeCode,
  'counterparty_type': _$CounterpartyTypeEnumMap[instance.counterpartyType]!,
  'amount_cents': instance.amountCents,
  'iva_rate': _$IvaRateEnumMap[instance.ivaRate]!,
  'iva_cents': instance.ivaCents,
  'id': ?instance.id,
  'counterparty_kind': ?_$CounterpartyKindEnumMap[instance.counterpartyKind],
  'operation_kind': ?_$OperationKindEnumMap[instance.operationKind],
  'residential_rent': ?instance.residentialRent,
  'via_platform': ?instance.viaPlatform,
  'isr_withheld_cents': ?instance.isrWithheldCents,
  'iva_withheld_cents': ?instance.ivaWithheldCents,
  'invoiced_on': ?instance.invoicedOn,
  'collected_on': ?instance.collectedOn,
  'platform_activity': ?_$PlatformActivityEnumMap[instance.platformActivity],
  'cfdi_uuid': ?instance.cfdiUuid,
  'description': ?instance.description,
};

const _$CounterpartyTypeEnumMap = {
  CounterpartyType.personaFisica: 'persona_fisica',
  CounterpartyType.personaMoral: 'persona_moral',
  CounterpartyType.publicoGeneral: 'publico_general',
  CounterpartyType.plataforma: 'plataforma',
  CounterpartyType.$unknown: r'$unknown',
};

const _$IvaRateEnumMap = {
  IvaRate.rate16: 'rate_16',
  IvaRate.rate8: 'rate_8',
  IvaRate.rate0: 'rate_0',
  IvaRate.exempt: 'exempt',
  IvaRate.$unknown: r'$unknown',
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

const _$PlatformActivityEnumMap = {
  PlatformActivity.transportDelivery: 'transport_delivery',
  PlatformActivity.lodging: 'lodging',
  PlatformActivity.goodsServices: 'goods_services',
  PlatformActivity.$unknown: r'$unknown',
};
