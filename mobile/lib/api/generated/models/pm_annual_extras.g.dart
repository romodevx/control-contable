// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'pm_annual_extras.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_PmAnnualExtras _$PmAnnualExtrasFromJson(Map<String, dynamic> json) =>
    _PmAnnualExtras(
      ptuBaseCents: (json['ptu_base_cents'] as num).toInt(),
      ptuToDistributeCents: (json['ptu_to_distribute_cents'] as num?)?.toInt(),
      suggestedCuNextYear: json['suggested_cu_next_year'] as String?,
    );

Map<String, dynamic> _$PmAnnualExtrasToJson(_PmAnnualExtras instance) =>
    <String, dynamic>{
      'ptu_base_cents': instance.ptuBaseCents,
      'ptu_to_distribute_cents': instance.ptuToDistributeCents,
      'suggested_cu_next_year': instance.suggestedCuNextYear,
    };
