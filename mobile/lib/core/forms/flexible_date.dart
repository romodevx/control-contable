/// Day, month and year separated by `/`, `-`, `.` or spaces (es-MX order).
final _separated = RegExp(r'^(\d{1,2})[/\-. ]+(\d{1,2})[/\-. ]+(\d{2}|\d{4})$');

/// Only digits: `ddMMyyyy` or `ddMMyy`.
final _compact = RegExp(r'^(\d{2})(\d{2})(\d{4}|\d{2})$');

/// ISO 8601 calendar date: `yyyy-MM-dd`.
final _iso = RegExp(r'^(\d{4})-(\d{1,2})-(\d{1,2})$');

/// Two-digit years are read in this century.
const _century = 2000;

/// The date in [text], or `null` when it is not a real calendar date.
///
/// Reads es-MX day-month-year with any common separator or none
/// (`31/12/2026`, `31-12-26`, `31122026`) and ISO `2026-12-31`. Never reads
/// month-first: `12/31/2026` is rejected, not guessed.
DateTime? parseFlexibleDate(String text) {
  final input = text.trim();
  final iso = _iso.firstMatch(input);
  if (iso != null) return _date(iso.group(3)!, iso.group(2)!, iso.group(1)!);
  final match = _separated.firstMatch(input) ?? _compact.firstMatch(input);
  if (match == null) return null;
  return _date(match.group(1)!, match.group(2)!, match.group(3)!);
}

/// [date] as `dd/MM/yyyy`, the form the date field shows.
String formatDate(DateTime date) =>
    '${_twoDigits(date.day)}/${_twoDigits(date.month)}/${date.year}';

DateTime? _date(String day, String month, String year) {
  final d = int.parse(day);
  final m = int.parse(month);
  final shortYear = year.length == 2;
  final y = int.parse(year) + (shortYear ? _century : 0);
  final date = DateTime(y, m, d);
  // DateTime rolls 31/02 over into March; a real date round-trips.
  final isReal = date.year == y && date.month == m && date.day == d;
  return isReal ? date : null;
}

String _twoDigits(int value) => value.toString().padLeft(2, '0');
