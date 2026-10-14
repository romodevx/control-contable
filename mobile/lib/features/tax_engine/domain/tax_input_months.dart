import 'package:control_contable/features/taxpayers/domain/periods.dart';

const int _monthsPerYear = 12;

/// Months shown at once in the monthly grid (ui-forms-and-modals.md: a
/// form section stays within six visible fields).
const int monthsPerGroup = 6;

/// The months of [fiscalYear] whose provisional payment can already be
/// known in [current] (`YYYY-MM`, America/Mexico_City): a past year's
/// twelve, or this year's months before the current one (a month is paid
/// in the next). A month in [stored] always shows, so no saved value is
/// hidden (and silently kept or dropped by the replace).
List<int> paymentMonths({
  required int fiscalYear,
  required String current,
  Set<int> stored = const {},
}) {
  final now = periodMonth(current);
  final lastDue = fiscalYear < now.year ? _monthsPerYear : now.month - 1;
  return {
    for (var month = 1; month <= lastDue; month++) month,
    ...stored,
  }.toList()..sort();
}

/// [months] (ascending) split by half of the year, January to June and
/// July to December, so the grid never shows more than [monthsPerGroup]
/// fields; a half with no month has no group.
List<List<int>> monthGroups(List<int> months) => [
  for (final half in [
    months.where((month) => month <= monthsPerGroup).toList(),
    months.where((month) => month > monthsPerGroup).toList(),
  ])
    if (half.isNotEmpty) half,
];
