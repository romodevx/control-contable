// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'filed_payment_entry.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_FiledPaymentEntry _$FiledPaymentEntryFromJson(Map<String, dynamic> json) =>
    _FiledPaymentEntry(
      month: (json['month'] as num).toInt(),
      amountCents: (json['amount_cents'] as num).toInt(),
    );

Map<String, dynamic> _$FiledPaymentEntryToJson(_FiledPaymentEntry instance) =>
    <String, dynamic>{
      'month': instance.month,
      'amount_cents': instance.amountCents,
    };
