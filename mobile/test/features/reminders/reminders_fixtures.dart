import 'package:control_contable/api/generated/models/reminder_settings.dart';
import 'package:control_contable/api/generated/models/tax_calendar.dart';

/// `GET /v1/reminder-settings` of a user who never saved.
const Map<String, Object?> defaultSettingsJson = {
  'enabled': false,
  'early_day': 14,
  'on_due_date': true,
  'send_time': '10:00',
  'show_amounts': false,
};

/// [defaultSettingsJson], parsed.
const ReminderSettings defaultSettings = ReminderSettings(
  enabled: false,
  earlyDay: 14,
  onDueDate: true,
  sendTime: '10:00',
  showAmounts: false,
);

/// A registered `Device` (never carries the token).
Map<String, Object?> deviceJson(String id) => {
  'id': id,
  'platform': 'android',
  'app_version': '0.1.0',
  'created_at': '2026-10-15T16:00:00Z',
  'last_seen_at': '2026-10-15T16:00:00Z',
};

/// The 12 entries of [year]: day 17 of the next month, with May 2026's
/// (period 2026-04) moved to Monday 18 by the weekend, as the server
/// answers. Not a calendar oracle: only the shape matters.
Map<String, Object?> taxCalendarJson(int year) => {
  'year': year,
  'entries': [
    for (var month = 1; month <= 12; month++) _entryJson(year, month),
  ],
};

/// [taxCalendarJson], parsed.
TaxCalendar taxCalendar(int year) =>
    TaxCalendar.fromJson(taxCalendarJson(year));

Map<String, Object?> _entryJson(int year, int month) {
  final dueYear = month == 12 ? year + 1 : year;
  final dueMonth = month == 12 ? 1 : month + 1;
  final nominal = '$dueYear-${_two(dueMonth)}-17';
  final moved = year == 2026 && month == 4;
  return {
    'period': '$year-${_two(month)}',
    'obligation': 'monthly_isr_iva',
    'nominal_due_date': nominal,
    'due_date': moved ? '$dueYear-${_two(dueMonth)}-18' : nominal,
    'moved_reason': moved ? 'weekend' : null,
    'subject_to_confirmation': true,
  };
}

String _two(int value) => value.toString().padLeft(2, '0');

/// The single error envelope with one detail [detailCode] on `body`.
Map<String, Object?> errorJson(String code, String detailCode) => {
  'error': {
    'code': code,
    'message': 'Refused.',
    'request_id': 'req_1',
    'details': [
      {'field': 'body', 'code': detailCode, 'message': 'Refused.'},
    ],
  },
};
