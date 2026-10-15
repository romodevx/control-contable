// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'filed_payments_update.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_FiledPaymentsUpdate _$FiledPaymentsUpdateFromJson(Map<String, dynamic> json) =>
    _FiledPaymentsUpdate(
      months: (json['months'] as List<dynamic>)
          .map((e) => FiledPaymentEntry.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$FiledPaymentsUpdateToJson(
  _FiledPaymentsUpdate instance,
) => <String, dynamic>{'months': instance.months};
