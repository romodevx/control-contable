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
      isrWithholdingAllowed: json['isr_withholding_allowed'] as bool,
      isrWithholdingRate: json['isr_withholding_rate'] as String?,
      ivaWithholdingAllowed: json['iva_withholding_allowed'] as bool,
      ivaWithholdingRate: json['iva_withholding_rate'] as String?,
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
      'isr_withholding_allowed': instance.isrWithholdingAllowed,
      'isr_withholding_rate': instance.isrWithholdingRate,
      'iva_withholding_allowed': instance.ivaWithholdingAllowed,
      'iva_withholding_rate': instance.ivaWithholdingRate,
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
