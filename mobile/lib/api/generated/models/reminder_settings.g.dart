// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'reminder_settings.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_ReminderSettings _$ReminderSettingsFromJson(Map<String, dynamic> json) =>
    _ReminderSettings(
      enabled: json['enabled'] as bool,
      earlyDay: (json['early_day'] as num?)?.toInt(),
      onDueDate: json['on_due_date'] as bool,
      sendTime: json['send_time'] as String,
      showAmounts: json['show_amounts'] as bool,
    );

Map<String, dynamic> _$ReminderSettingsToJson(_ReminderSettings instance) =>
    <String, dynamic>{
      'enabled': instance.enabled,
      'early_day': instance.earlyDay,
      'on_due_date': instance.onDueDate,
      'send_time': instance.sendTime,
      'show_amounts': instance.showAmounts,
    };
