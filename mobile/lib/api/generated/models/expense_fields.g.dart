// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'expense_fields.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_ExpenseFields _$ExpenseFieldsFromJson(Map<String, dynamic> json) =>
    _ExpenseFields(
      regimeCode: json['regime_code'] as String?,
      categoryCode: json['category_code'] as String?,
      description: json['description'] as String?,
      amountCents: (json['amount_cents'] as num?)?.toInt(),
      ivaRate: json['iva_rate'] == null
          ? null
          : IvaRate.fromJson(json['iva_rate'] as String),
      ivaCents: (json['iva_cents'] as num?)?.toInt(),
      paymentMethod: json['payment_method'] == null
          ? null
          : PaymentMethod.fromJson(json['payment_method'] as String),
      hasCfdi: json['has_cfdi'] as bool?,
      paidOn: json['paid_on'] == null
          ? null
          : DateTime.parse(json['paid_on'] as String),
      isrWithheldCents: (json['isr_withheld_cents'] as num?)?.toInt(),
      ivaWithheldCents: (json['iva_withheld_cents'] as num?)?.toInt(),
    );

Map<String, dynamic> _$ExpenseFieldsToJson(_ExpenseFields instance) =>
    <String, dynamic>{
      'regime_code': instance.regimeCode,
      'category_code': instance.categoryCode,
      'description': instance.description,
      'amount_cents': instance.amountCents,
      'iva_rate': _$IvaRateEnumMap[instance.ivaRate],
      'iva_cents': instance.ivaCents,
      'payment_method': _$PaymentMethodEnumMap[instance.paymentMethod],
      'has_cfdi': instance.hasCfdi,
      'paid_on': instance.paidOn?.toIso8601String(),
      'isr_withheld_cents': instance.isrWithheldCents,
      'iva_withheld_cents': instance.ivaWithheldCents,
    };

const _$IvaRateEnumMap = {
  IvaRate.rate16: 'rate_16',
  IvaRate.rate8: 'rate_8',
  IvaRate.rate0: 'rate_0',
  IvaRate.exempt: 'exempt',
  IvaRate.$unknown: r'$unknown',
};

const _$PaymentMethodEnumMap = {
  PaymentMethod.cash: 'cash',
  PaymentMethod.transfer: 'transfer',
  PaymentMethod.cardDebit: 'card_debit',
  PaymentMethod.cardCredit: 'card_credit',
  PaymentMethod.check: 'check',
  PaymentMethod.eWallet: 'e_wallet',
  PaymentMethod.other: 'other',
  PaymentMethod.$unknown: r'$unknown',
};
