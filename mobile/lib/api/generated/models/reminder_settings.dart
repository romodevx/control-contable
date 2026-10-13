// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

part 'reminder_settings.freezed.dart';
part 'reminder_settings.g.dart';

@Freezed()
abstract class ReminderSettings with _$ReminderSettings {
  const factory ReminderSettings({
    @Default(false) bool enabled,

    /// Day of month for the early reminder; null disables it.
    @JsonKey(includeIfNull: true, name: 'early_day') @Default(14) int? earlyDay,
    @JsonKey(name: 'on_due_date') @Default(true) bool onDueDate,

    /// Local time, America/Mexico_City.
    @JsonKey(name: 'send_time') @Default('10:00') String sendTime,

    /// Lock-screen privacy opt-in.
    @JsonKey(name: 'show_amounts') @Default(false) bool showAmounts,
  }) = _ReminderSettings;

  factory ReminderSettings.fromJson(Map<String, Object?> json) =>
      _$ReminderSettingsFromJson(json);
}
