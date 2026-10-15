/// The first fiscal year with published tax parameters (MODULE-history
/// §2.1): no history exists before it.
const int firstHistoryYear = 2026;

/// The fiscal years the picker offers, newest first: [currentYear] back to
/// [firstHistoryYear]. Always offers [firstHistoryYear].
List<int> historyYears(int currentYear) => [
  for (var year = currentYear; year >= firstHistoryYear; year--) year,
  if (currentYear < firstHistoryYear) firstHistoryYear,
];
