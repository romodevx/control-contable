import 'package:control_contable/api/generated/models/tax_calendar_entry.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/reminders/data/reminders_repository.dart';
import 'package:control_contable/features/reminders/domain/upcoming_due_dates.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The next due dates from today in Mexico City (MODULE-reminders §7).
///
/// A year's calendar lists its **periods**, so December's entry is due in
/// January of the next year: in January the previous year is read too, and
/// late in the year the next one. A year the server has not seeded
/// (`422 calendar_unavailable`) is simply left out.
final FutureProvider<List<TaxCalendarEntry>> upcomingDueDatesProvider =
    FutureProvider.autoDispose<List<TaxCalendarEntry>>(
      (ref) async {
        final repository = ref.read(remindersRepositoryProvider);
        final today = businessToday(ref.read(clockProvider)());
        final entries = [
          if (today.month == DateTime.january)
            ...await _entriesOrNone(repository, today.year - 1),
          ...(await repository.getCalendar(today.year)).entries,
        ];
        final next = upcomingDueDates(entries, today: today);
        if (next.length == upcomingDueDateCount) return next;
        return upcomingDueDates([
          ...entries,
          ...await _entriesOrNone(repository, today.year + 1),
        ], today: today);
      },
      name: 'upcomingDueDatesProvider',
      retry: (retryCount, error) => null,
    );

Future<List<TaxCalendarEntry>> _entriesOrNone(
  RemindersRepository repository,
  int year,
) async {
  try {
    return (await repository.getCalendar(year)).entries;
  } on ApiRejected {
    return const [];
  }
}
