import 'package:control_contable/features/taxpayers/domain/periods.dart';

/// The first month the server can estimate: January of the first fiscal
/// year with transcribed parameters (2026, MODULE-tax-engine OQ-T9). An
/// earlier month answers `422 parameters_unavailable`.
const String firstEstimablePeriod = '2026-01';

/// The months the estimate screen offers, newest first: from [current]
/// (the still-open month, America/Mexico_City) back to
/// [firstEstimablePeriod]. Future months are never offered (the server
/// answers `400 out_of_range`). Empty when [current] is before the first.
List<String> estimablePeriods({required String current}) {
  if (periodMonth(current).isBefore(periodMonth(firstEstimablePeriod))) {
    return const [];
  }
  return endPeriodChoices(start: firstEstimablePeriod, current: current);
}

/// The fiscal year (calendar year) of [period] (`YYYY-MM`).
int fiscalYearOf(String period) => periodMonth(period).year;
