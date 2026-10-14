/// Mexico City's offset from UTC (no daylight saving time since 2022).
const Duration _mexicoCityOffset = Duration(hours: -6);

const int _monthsPerYear = 12;
const int _monthDigits = 2;
const String _separator = '-';

/// The `YYYY-MM` period of [now] in America/Mexico_City (the server's
/// calendar for `start_period` and `end_period`).
String currentPeriod(DateTime now) {
  final local = now.toUtc().add(_mexicoCityOffset);
  return _format(local.year, local.month);
}

/// The first day of [period] (`YYYY-MM`).
DateTime periodMonth(String period) {
  final [year, month] = period.split(_separator).map(int.parse).toList();
  return DateTime(year, month);
}

/// The months a regime that started in [start] can end in, newest first:
/// from [current] back to [start] (the server allows
/// `start_period ≤ end_period ≤ current month`). Always offers [current].
List<String> endPeriodChoices({
  required String start,
  required String current,
}) {
  final first = periodMonth(start);
  var month = periodMonth(current);
  final choices = [current];
  while (true) {
    month = DateTime(month.year, month.month - 1);
    if (month.isBefore(first)) return choices;
    choices.add(_format(month.year, month.month));
  }
}

/// The `YYYY-MM` period of [month] (1-12) in [year].
String periodOf(int year, int month) => _format(year, month);

String _format(int year, int month) {
  assert(month >= 1 && month <= _monthsPerYear, 'month out of range');
  return '$year$_separator${month.toString().padLeft(_monthDigits, '0')}';
}
