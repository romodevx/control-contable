// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'expense_category_list.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_ExpenseCategoryList _$ExpenseCategoryListFromJson(Map<String, dynamic> json) =>
    _ExpenseCategoryList(
      items: (json['items'] as List<dynamic>)
          .map((e) => ExpenseCategory.fromJson(e as Map<String, dynamic>))
          .toList(),
      ivaRatesAllowed: (json['iva_rates_allowed'] as List<dynamic>)
          .map((e) => IvaRate.fromJson(e as String))
          .toList(),
      cashPaymentLimitCents: (json['cash_payment_limit_cents'] as num).toInt(),
      legend: json['legend'] as String,
    );

Map<String, dynamic> _$ExpenseCategoryListToJson(
  _ExpenseCategoryList instance,
) => <String, dynamic>{
  'items': instance.items,
  'iva_rates_allowed': instance.ivaRatesAllowed
      .map((e) => _$IvaRateEnumMap[e]!)
      .toList(),
  'cash_payment_limit_cents': instance.cashPaymentLimitCents,
  'legend': instance.legend,
};

const _$IvaRateEnumMap = {
  IvaRate.rate16: 'rate_16',
  IvaRate.rate8: 'rate_8',
  IvaRate.rate0: 'rate_0',
  IvaRate.exempt: 'exempt',
  IvaRate.$unknown: r'$unknown',
};
