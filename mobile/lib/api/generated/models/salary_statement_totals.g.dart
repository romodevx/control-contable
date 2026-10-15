// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'salary_statement_totals.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_SalaryStatementTotals _$SalaryStatementTotalsFromJson(
  Map<String, dynamic> json,
) => _SalaryStatementTotals(
  accumulableCents: (json['accumulable_cents'] as num).toInt(),
  exemptCents: (json['exempt_cents'] as num).toInt(),
  isrWithheldCents: (json['isr_withheld_cents'] as num).toInt(),
);

Map<String, dynamic> _$SalaryStatementTotalsToJson(
  _SalaryStatementTotals instance,
) => <String, dynamic>{
  'accumulable_cents': instance.accumulableCents,
  'exempt_cents': instance.exemptCents,
  'isr_withheld_cents': instance.isrWithheldCents,
};
