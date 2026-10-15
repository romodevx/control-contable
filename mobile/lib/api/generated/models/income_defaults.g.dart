// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'income_defaults.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_IncomeDefaults _$IncomeDefaultsFromJson(Map<String, dynamic> json) =>
    _IncomeDefaults(
      ivaRatesAllowed: (json['iva_rates_allowed'] as List<dynamic>)
          .map((e) => IvaRate.fromJson(e as String))
          .toList(),
      suggestedIvaRate: IvaRate.fromJson(json['suggested_iva_rate'] as String),
      lastUsed: IncomeLastUsed.fromJson(
        json['last_used'] as Map<String, dynamic>,
      ),
      platformActivity: json['platform_activity'] == null
          ? null
          : PlatformActivity.fromJson(json['platform_activity'] as String),
    );

Map<String, dynamic> _$IncomeDefaultsToJson(_IncomeDefaults instance) =>
    <String, dynamic>{
      'iva_rates_allowed': instance.ivaRatesAllowed
          .map((e) => _$IvaRateEnumMap[e]!)
          .toList(),
      'suggested_iva_rate': _$IvaRateEnumMap[instance.suggestedIvaRate]!,
      'last_used': instance.lastUsed,
      'platform_activity': _$PlatformActivityEnumMap[instance.platformActivity],
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
