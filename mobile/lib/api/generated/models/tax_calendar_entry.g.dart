// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'tax_calendar_entry.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_TaxCalendarEntry _$TaxCalendarEntryFromJson(Map<String, dynamic> json) =>
    _TaxCalendarEntry(
      period: json['period'] as String,
      obligation: TaxObligation.fromJson(json['obligation'] as String),
      nominalDueDate: json['nominal_due_date'] as String,
      dueDate: json['due_date'] as String,
      movedReason: json['moved_reason'] == null
          ? null
          : TaxCalendarMovedReason.fromJson(json['moved_reason'] as String),
      subjectToConfirmation: json['subject_to_confirmation'] as bool,
    );

Map<String, dynamic> _$TaxCalendarEntryToJson(_TaxCalendarEntry instance) =>
    <String, dynamic>{
      'period': instance.period,
      'obligation': _$TaxObligationEnumMap[instance.obligation]!,
      'nominal_due_date': instance.nominalDueDate,
      'due_date': instance.dueDate,
      'moved_reason': _$TaxCalendarMovedReasonEnumMap[instance.movedReason],
      'subject_to_confirmation': instance.subjectToConfirmation,
    };

const _$TaxObligationEnumMap = {
  TaxObligation.monthlyIsrIva: 'monthly_isr_iva',
  TaxObligation.$unknown: r'$unknown',
};

const _$TaxCalendarMovedReasonEnumMap = {
  TaxCalendarMovedReason.weekend: 'weekend',
  TaxCalendarMovedReason.holiday: 'holiday',
  TaxCalendarMovedReason.$unknown: r'$unknown',
};
