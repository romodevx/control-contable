// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'period.dart';
import 'tax_calendar_moved_reason.dart';
import 'tax_obligation.dart';

part 'tax_calendar_entry.freezed.dart';
part 'tax_calendar_entry.g.dart';

@Freezed()
abstract class TaxCalendarEntry with _$TaxCalendarEntry {
  const factory TaxCalendarEntry({
    required Period period,
    required TaxObligation obligation,

    /// Day 17 of the month after `period`.
    @JsonKey(name: 'nominal_due_date') required String nominalDueDate,

    /// The nominal date moved to the next business day when needed.
    @JsonKey(name: 'due_date') required String dueDate,

    /// `null` when `due_date` equals `nominal_due_date`.
    @JsonKey(includeIfNull: true, name: 'moved_reason')
    required TaxCalendarMovedReason? movedReason,

    /// `true` while the date rule awaits CPA confirmation (OQ-R1, the CFF 12 ¶5 Friday extension is not applied): today every entry. The app shows "sujeto a confirmación".
    @JsonKey(name: 'subject_to_confirmation')
    required bool subjectToConfirmation,
  }) = _TaxCalendarEntry;

  factory TaxCalendarEntry.fromJson(Map<String, Object?> json) =>
      _$TaxCalendarEntryFromJson(json);
}
