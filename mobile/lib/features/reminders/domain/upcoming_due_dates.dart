import 'package:control_contable/api/generated/models/tax_calendar_entry.dart';
import 'package:control_contable/api/generated/models/tax_obligation.dart';
import 'package:control_contable/core/records/record_month.dart';

/// How many due dates the reminders screen lists (MODULE-reminders §7).
const int upcomingDueDateCount = 3;

/// The first [count] entries of [entries] (in calendar order) due on or
/// after [today]. Entries of an obligation this build does not know are
/// skipped, as the contract asks.
List<TaxCalendarEntry> upcomingDueDates(
  Iterable<TaxCalendarEntry> entries, {
  required DateTime today,
  int count = upcomingDueDateCount,
}) {
  final todayIso = businessDate(today);
  return entries
      .where((entry) => entry.obligation != TaxObligation.$unknown)
      .where((entry) => entry.dueDate.compareTo(todayIso) >= 0)
      .take(count)
      .toList();
}
