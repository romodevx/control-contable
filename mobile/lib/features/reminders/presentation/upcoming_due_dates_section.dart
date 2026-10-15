import 'package:control_contable/api/generated/models/tax_calendar_entry.dart';
import 'package:control_contable/api/generated/models/tax_calendar_moved_reason.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/reminders/application/upcoming_due_dates.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// "Próximos vencimientos": the next due dates, the moved ones marked, and
/// the "sujeto a confirmación" note (OQ-R1, MODULE-reminders §7).
class UpcomingDueDatesSection extends ConsumerWidget {
  /// Creates the section.
  const new({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final upcoming = ref.watch(upcomingDueDatesProvider);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Semantics(
          header: true,
          child: Text(
            l10n.remindersUpcomingTitle,
            style: Theme.of(context).textTheme.titleMedium,
          ),
        ),
        const SizedBox(height: AppSpacing.xs),
        switch (upcoming) {
          AsyncData(:final value) when value.isEmpty => LoadStatus(
            message: l10n.remindersUpcomingEmpty,
          ),
          AsyncData(:final value) => _DueDateList(entries: value),
          AsyncError() => LoadFailed(
            message: l10n.remindersUpcomingFailed,
            onRetry: () => ref.invalidate(upcomingDueDatesProvider),
          ),
          _ => LoadStatus(message: l10n.remindersUpcomingLoading, busy: true),
        },
      ],
    );
  }
}

class _DueDateList extends StatelessWidget {
  const new({required this.entries});

  final List<TaxCalendarEntry> entries;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        for (final entry in entries) _DueDateRow(entry: entry),
        if (entries.any((entry) => entry.subjectToConfirmation))
          Text(
            l10n.remindersSubjectToConfirmation,
            style: Theme.of(context).textTheme.bodySmall,
          ),
      ],
    );
  }
}

class _DueDateRow extends StatelessWidget {
  const new({required this.entry});

  final TaxCalendarEntry entry;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final textTheme = Theme.of(context).textTheme;
    final moved = switch (entry.movedReason) {
      TaxCalendarMovedReason.weekend => l10n.remindersMovedWeekend,
      TaxCalendarMovedReason.holiday => l10n.remindersMovedHoliday,
      _ => null,
    };
    return Padding(
      padding: const EdgeInsets.only(bottom: AppSpacing.sm),
      child: MergeSemantics(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              l10n.remindersUpcomingDate(parseBusinessDate(entry.dueDate)),
              style: textTheme.bodyLarge,
            ),
            Text(
              l10n.remindersUpcomingPeriod(_periodMonth(entry.period)),
              style: textTheme.bodyMedium,
            ),
            if (moved != null) Text(moved, style: textTheme.bodySmall),
          ],
        ),
      ),
    );
  }

  static DateTime _periodMonth(String period) =>
      parseBusinessDate('$period-01');
}
