// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'income.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Income _$IncomeFromJson(Map<String, dynamic> json) => _Income(
  id: json['id'] as String,
  taxpayerId: json['taxpayer_id'] as String,
  regimeCode: json['regime_code'] as String,
  counterpartyType: CounterpartyType.fromJson(
    json['counterparty_type'] as String,
  ),
  amountCents: (json['amount_cents'] as num).toInt(),
  ivaRate: IvaRate.fromJson(json['iva_rate'] as String),
  ivaCents: (json['iva_cents'] as num).toInt(),
  isrWithheldCents: (json['isr_withheld_cents'] as num).toInt(),
  ivaWithheldCents: (json['iva_withheld_cents'] as num).toInt(),
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
  'counterparty_type': _$CounterpartyTypeEnumMap[instance.counterpartyType]!,
  'amount_cents': instance.amountCents,
  'iva_rate': _$IvaRateEnumMap[instance.ivaRate]!,
  'iva_cents': instance.ivaCents,
  'isr_withheld_cents': instance.isrWithheldCents,
  'iva_withheld_cents': instance.ivaWithheldCents,
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

const _$PlatformActivityEnumMap = {
  PlatformActivity.transportDelivery: 'transport_delivery',
  PlatformActivity.lodging: 'lodging',
  PlatformActivity.goodsServices: 'goods_services',
  PlatformActivity.$unknown: r'$unknown',
};
