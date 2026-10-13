// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'iva_components.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_IvaComponents _$IvaComponentsFromJson(Map<String, dynamic> json) =>
    _IvaComponents(
      chargedCollectedCents: (json['charged_collected_cents'] as num).toInt(),
      creditablePaidCents: (json['creditable_paid_cents'] as num).toInt(),
      withheldFromTaxpayerCents: (json['withheld_from_taxpayer_cents'] as num)
          .toInt(),
      definitiveExcludedCents: (json['definitive_excluded_cents'] as num?)
          ?.toInt(),
    );

Map<String, dynamic> _$IvaComponentsToJson(_IvaComponents instance) =>
    <String, dynamic>{
      'charged_collected_cents': instance.chargedCollectedCents,
      'creditable_paid_cents': instance.creditablePaidCents,
      'withheld_from_taxpayer_cents': instance.withheldFromTaxpayerCents,
      'definitive_excluded_cents': instance.definitiveExcludedCents,
    };
