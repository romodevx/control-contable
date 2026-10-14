// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'period.dart';
import 'tax_calendar_entry_moved_reason.dart';
import 'tax_calendar_entry_obligation.dart';

part 'tax_calendar_entry.freezed.dart';
part 'tax_calendar_entry.g.dart';

@Freezed()
abstract class TaxCalendarEntry with _$TaxCalendarEntry {
  const factory TaxCalendarEntry({
    required Period period,
    required TaxCalendarEntryObligation obligation,
    @JsonKey(name: 'nominal_due_date') required String nominalDueDate,
    @JsonKey(name: 'due_date') required String dueDate,
    @JsonKey(includeIfNull: false, name: 'moved_reason')
    TaxCalendarEntryMovedReason? movedReason,
  }) = _TaxCalendarEntry;

  factory TaxCalendarEntry.fromJson(Map<String, Object?> json) =>
      _$TaxCalendarEntryFromJson(json);
}
