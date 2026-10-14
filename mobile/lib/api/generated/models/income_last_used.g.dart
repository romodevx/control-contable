// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'income_last_used.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_IncomeLastUsed _$IncomeLastUsedFromJson(Map<String, dynamic> json) =>
    _IncomeLastUsed(
      counterpartyKind: json['counterparty_kind'] == null
          ? null
          : CounterpartyKind.fromJson(json['counterparty_kind'] as String),
      operationKind: OperationKind.fromJson(json['operation_kind'] as String),
      residentialRent: json['residential_rent'] as bool,
      viaPlatform: json['via_platform'] as bool,
    );

Map<String, dynamic> _$IncomeLastUsedToJson(_IncomeLastUsed instance) =>
    <String, dynamic>{
      'counterparty_kind': _$CounterpartyKindEnumMap[instance.counterpartyKind],
      'operation_kind': _$OperationKindEnumMap[instance.operationKind]!,
      'residential_rent': instance.residentialRent,
      'via_platform': instance.viaPlatform,
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
