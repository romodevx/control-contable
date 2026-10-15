// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'fiscal_year.dart';
import 'tax_calendar_entry.dart';

part 'tax_calendar.freezed.dart';
part 'tax_calendar.g.dart';

@Freezed()
abstract class TaxCalendar with _$TaxCalendar {
  const factory TaxCalendar({
    required FiscalYear year,

    /// Periods `year-01` .. `year-12`, in order.
    required List<TaxCalendarEntry> entries,
  }) = _TaxCalendar;

  factory TaxCalendar.fromJson(Map<String, Object?> json) =>
      _$TaxCalendarFromJson(json);
}
