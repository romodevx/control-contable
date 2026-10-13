// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'income_create.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_IncomeCreate _$IncomeCreateFromJson(Map<String, dynamic> json) =>
    _IncomeCreate(
      regimeCode: json['regime_code'] as String?,
      description: json['description'] as String?,
      counterpartyType: json['counterparty_type'] == null
          ? null
          : CounterpartyType.fromJson(json['counterparty_type'] as String),
      amountCents: (json['amount_cents'] as num?)?.toInt(),
      ivaRate: json['iva_rate'] == null
          ? null
          : IvaRate.fromJson(json['iva_rate'] as String),
      ivaCents: (json['iva_cents'] as num?)?.toInt(),
      isrWithheldCents: (json['isr_withheld_cents'] as num?)?.toInt(),
      ivaWithheldCents: (json['iva_withheld_cents'] as num?)?.toInt(),
      invoicedOn: json['invoiced_on'] == null
          ? null
          : DateTime.parse(json['invoiced_on'] as String),
      collectedOn: json['collected_on'] == null
          ? null
          : DateTime.parse(json['collected_on'] as String),
      id: json['id'] as String?,
    );

Map<String, dynamic> _$IncomeCreateToJson(
  _IncomeCreate instance,
) => <String, dynamic>{
  'regime_code': ?instance.regimeCode,
  'description': ?instance.description,
  'counterparty_type': ?_$CounterpartyTypeEnumMap[instance.counterpartyType],
  'amount_cents': ?instance.amountCents,
  'iva_rate': ?_$IvaRateEnumMap[instance.ivaRate],
  'iva_cents': ?instance.ivaCents,
  'isr_withheld_cents': ?instance.isrWithheldCents,
  'iva_withheld_cents': ?instance.ivaWithheldCents,
  'invoiced_on': ?instance.invoicedOn?.toIso8601String(),
  'collected_on': ?instance.collectedOn?.toIso8601String(),
  'id': ?instance.id,
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
