import 'package:flutter/foundation.dart';

/// Mexico City's offset from UTC (no daylight saving time since 2022): the
/// server's calendar for record dates and periods.
const Duration _mexicoCityOffset = Duration(hours: -6);

const int _monthsPerYear = 12;
const int _monthDigits = 2;
const int _dayDigits = 2;
const int _yearDigits = 4;
const String _separator = '-';

/// A `YYYY-MM` month as a location carries it.
final RegExp _periodPattern = RegExp(r'^\d{4}-(0[1-9]|1[0-2])$');

/// The query key a record list location names its month with.
const String recordMonthQuery = 'periodo';

/// A calendar month of records (incomes, expenses): the `period` query of
/// the list endpoints (`YYYY-MM`).
@immutable
class RecordMonth {
  /// The month [month] (1–12) of [year].
  const new(this.year, this.month)
    : assert(month >= 1 && month <= _monthsPerYear, 'month out of range');

  /// The month of a `YYYY-MM` [period].
  factory parse(String period) {
    final [year, month] = period.split(_separator).map(int.parse).toList();
    return RecordMonth(year, month);
  }

  /// The month a calendar [date] falls in.
  factory ofDate(DateTime date) => RecordMonth(date.year, date.month);

  /// The current month in America/Mexico_City at [now].
  factory current(DateTime now) => RecordMonth.ofDate(businessToday(now));

  /// The month a list opens on: [period] when it is a `YYYY-MM` month no
  /// later than the current one at [now], else the current month (a
  /// location is user input).
  factory opening(String? period, DateTime now) {
    final current = RecordMonth.current(now);
    if (period == null || !_periodPattern.hasMatch(period)) return current;
    final named = RecordMonth.parse(period);
    return current.isBefore(named) ? current : named;
  }

  /// The year.
  final int year;

  /// The month, 1–12.
  final int month;

  /// `YYYY-MM`.
  String get period =>
      '${_pad(year, _yearDigits)}$_separator${_pad(month, _monthDigits)}';

  /// The first day of the month (local calendar date).
  DateTime get firstDay => DateTime(year, month);

  /// The month before.
  RecordMonth get previous => RecordMonth.ofDate(DateTime(year, month - 1));

  /// The month after.
  RecordMonth get next => RecordMonth.ofDate(DateTime(year, month + 1));

  /// Whether this month comes before [other].
  bool isBefore(RecordMonth other) =>
      year < other.year || (year == other.year && month < other.month);

  @override
  bool operator ==(Object other) =>
      other is RecordMonth && other.year == year && other.month == month;

  @override
  int get hashCode => Object.hash(year, month);

  @override
  String toString() => period;
}

/// Today's calendar date in America/Mexico_City at [now] (the server's
/// bound for record dates).
DateTime businessToday(DateTime now) {
  final local = now.toUtc().add(_mexicoCityOffset);
  return DateTime(local.year, local.month, local.day);
}

/// [date] as the contract's `BusinessDate` (`YYYY-MM-DD`).
String businessDate(DateTime date) =>
    '${_pad(date.year, _yearDigits)}$_separator'
    '${_pad(date.month, _monthDigits)}$_separator'
    '${_pad(date.day, _dayDigits)}';

/// The calendar date of a `BusinessDate` (`YYYY-MM-DD`).
DateTime parseBusinessDate(String date) {
  final [year, month, day] = date.split(_separator).map(int.parse).toList();
  return DateTime(year, month, day);
}

String _pad(int value, int width) => value.toString().padLeft(width, '0');
