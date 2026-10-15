// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'annual_balance.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_AnnualBalance _$AnnualBalanceFromJson(Map<String, dynamic> json) =>
    _AnnualBalance(
      kind: AnnualBalanceKind.fromJson(json['kind'] as String),
      amountCents: (json['amount_cents'] as num).toInt(),
    );

Map<String, dynamic> _$AnnualBalanceToJson(_AnnualBalance instance) =>
    <String, dynamic>{
      'kind': _$AnnualBalanceKindEnumMap[instance.kind]!,
      'amount_cents': instance.amountCents,
    };

const _$AnnualBalanceKindEnumMap = {
  AnnualBalanceKind.aCargo: 'a_cargo',
  AnnualBalanceKind.aFavor: 'a_favor',
  AnnualBalanceKind.cero: 'cero',
  AnnualBalanceKind.$unknown: r'$unknown',
};
