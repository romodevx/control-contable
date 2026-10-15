/// The first fiscal year with transcribed tax parameters: the server
/// refuses an annual estimate before it (`parameters_unavailable`).
const int firstAnnualYear = 2026;

/// Mexico City's offset from UTC (no daylight saving time since 2022).
const Duration _mexicoCityOffset = Duration(hours: -6);

/// The last month in which the previous year is still the one being
/// filed: personas físicas file it until 30 April (MODULE-annual OQ-A8).
const int _lastFilingMonth = DateTime.april;

/// The fiscal year the annual screen opens on at [now]: the previous year
/// from January to 30 April (America/Mexico_City), the current year after;
/// never before [firstAnnualYear].
int defaultAnnualYear(DateTime now) {
  final local = _inMexicoCity(now);
  final year = local.month <= _lastFilingMonth ? local.year - 1 : local.year;
  return year < firstAnnualYear ? firstAnnualYear : year;
}

/// The fiscal years the picker offers at [now], newest first: the current
/// year back to [firstAnnualYear]. Always offers [firstAnnualYear].
List<int> annualYears(DateTime now) {
  final current = _inMexicoCity(now).year;
  return [
    for (var year = current; year >= firstAnnualYear; year--) year,
    if (current < firstAnnualYear) firstAnnualYear,
  ];
}

DateTime _inMexicoCity(DateTime now) => now.toUtc().add(_mexicoCityOffset);
