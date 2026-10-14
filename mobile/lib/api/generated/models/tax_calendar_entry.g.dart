// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'tax_calendar_entry.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_TaxCalendarEntry _$TaxCalendarEntryFromJson(
  Map<String, dynamic> json,
) => _TaxCalendarEntry(
  period: json['period'] as String,
  obligation: TaxCalendarEntryObligation.fromJson(json['obligation'] as String),
  nominalDueDate: json['nominal_due_date'] as String,
  dueDate: json['due_date'] as String,
  movedReason: json['moved_reason'] == null
      ? null
      : TaxCalendarEntryMovedReason.fromJson(json['moved_reason'] as String?),
);

Map<String, dynamic> _$TaxCalendarEntryToJson(
  _TaxCalendarEntry instance,
) => <String, dynamic>{
  'period': instance.period,
  'obligation': _$TaxCalendarEntryObligationEnumMap[instance.obligation]!,
  'nominal_due_date': instance.nominalDueDate,
  'due_date': instance.dueDate,
  'moved_reason': ?_$TaxCalendarEntryMovedReasonEnumMap[instance.movedReason],
};

const _$TaxCalendarEntryObligationEnumMap = {
  TaxCalendarEntryObligation.monthlyIsrIva: 'monthly_isr_iva',
  TaxCalendarEntryObligation.$unknown: r'$unknown',
};

const _$TaxCalendarEntryMovedReasonEnumMap = {
  TaxCalendarEntryMovedReason.weekend: 'weekend',
  TaxCalendarEntryMovedReason.holiday: 'holiday',
  TaxCalendarEntryMovedReason.valueNull: null,
  TaxCalendarEntryMovedReason.$unknown: r'$unknown',
};
