// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'taxpayer_list.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_TaxpayerList _$TaxpayerListFromJson(Map<String, dynamic> json) =>
    _TaxpayerList(
      items: (json['items'] as List<dynamic>)
          .map((e) => Taxpayer.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$TaxpayerListToJson(_TaxpayerList instance) =>
    <String, dynamic>{'items': instance.items};
