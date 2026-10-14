// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'income_page.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_IncomePage _$IncomePageFromJson(Map<String, dynamic> json) => _IncomePage(
  items: (json['items'] as List<dynamic>)
      .map((e) => Income.fromJson(e as Map<String, dynamic>))
      .toList(),
  nextCursor: json['next_cursor'] as String?,
  totals: PeriodTotals.fromJson(json['totals'] as Map<String, dynamic>),
);

Map<String, dynamic> _$IncomePageToJson(_IncomePage instance) =>
    <String, dynamic>{
      'items': instance.items,
      'next_cursor': instance.nextCursor,
      'totals': instance.totals,
    };
