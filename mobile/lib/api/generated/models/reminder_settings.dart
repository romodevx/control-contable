// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

part 'reminder_settings.freezed.dart';
part 'reminder_settings.g.dart';

/// The caller's reminder preferences (MODULE-reminders §2.2), the body of both `getReminderSettings` and `putReminderSettings`. Every field is required, `early_day` included (nullable): PUT replaces the whole object, so nothing is defaulted from an omission. Defaults, answered by GET until the first save: `enabled: false`, `early_day: 14`, `on_due_date: true`, `send_time: 10:00`, `show_amounts: false`.
@Freezed()
abstract class ReminderSettings with _$ReminderSettings {
  const factory ReminderSettings({
    /// Opt-in. The app asks for the OS notification permission before sending `true`.
    required bool enabled,

    /// Day of the due month for the early reminder (nominal calendar day, sent even on a weekend); `null` = no early reminder.
    @JsonKey(includeIfNull: true, name: 'early_day') required int? earlyDay,

    /// Remind on the due date itself (after any weekend or holiday move).
    @JsonKey(name: 'on_due_date') required bool onDueDate,

    /// `HH:MM`, America/Mexico_City (every user, no per-user zone); delivered within one reminder-job tick (5 minutes).
    @JsonKey(name: 'send_time') required String sendTime,

    /// Show each taxpayer's estimated amount in the notification. Privacy opt-in: the OS shows it on the lock screen. Without it, or with `module_tax_engine` OFF, the notification lists names only.
    @JsonKey(name: 'show_amounts') required bool showAmounts,
  }) = _ReminderSettings;

  factory ReminderSettings.fromJson(Map<String, Object?> json) =>
      _$ReminderSettingsFromJson(json);
}
