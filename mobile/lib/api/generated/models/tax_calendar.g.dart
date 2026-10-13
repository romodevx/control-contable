// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'tax_calendar.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_TaxCalendar _$TaxCalendarFromJson(Map<String, dynamic> json) => _TaxCalendar(
  year: (json['year'] as num).toInt(),
  entries: (json['entries'] as List<dynamic>)
      .map((e) => TaxCalendarEntry.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$TaxCalendarToJson(_TaxCalendar instance) =>
    <String, dynamic>{'year': instance.year, 'entries': instance.entries};
