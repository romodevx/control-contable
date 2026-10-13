// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'regime_list.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_RegimeList _$RegimeListFromJson(Map<String, dynamic> json) => _RegimeList(
  items: (json['items'] as List<dynamic>)
      .map((e) => Regime.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$RegimeListToJson(_RegimeList instance) =>
    <String, dynamic>{'items': instance.items};
