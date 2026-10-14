import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/history/application/history.dart';
import 'package:control_contable/features/history/domain/history_years.dart';
import 'package:control_contable/features/history/presentation/history_states.dart';
import 'package:control_contable/features/history/presentation/history_year_picker.dart';
import 'package:control_contable/features/history/presentation/history_year_view.dart';
import 'package:control_contable/features/taxpayers/application/active_taxpayer.dart';
import 'package:control_contable/features/taxpayers/domain/periods.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// History of the active taxpayer (M8-05): a fiscal-year picker (the
/// current year by default), then every month of it, newest first, with
/// its totals, status and changed marker. Pull down to read it again.
class HistoryScreen extends ConsumerWidget {
  /// Creates the screen.
  const new({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final currentYear = periodMonth(currentPeriod(ref.watch(clockProvider)()))
        .year;
    final year = ref.watch(historyChosenYearProvider) ?? currentYear;
    final taxpayer = ref.watch(activeTaxpayerProvider);
    final taxpayerId = taxpayer.value?.id;
    final query = taxpayerId == null
        ? null
        : (taxpayerId: taxpayerId, fiscalYear: year);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.historyTitle)),
      body: RefreshIndicator(
        onRefresh: () => _refresh(ref, query),
        child: ListView(
          padding: const EdgeInsets.all(AppSpacing.gutter),
          physics: const AlwaysScrollableScrollPhysics(),
          children: [
            HistoryYearPicker(
              years: historyYears(currentYear),
              selected: year,
              onSelected: (chosen) => ref
                  .read(historyChosenYearProvider.notifier)
                  .select(chosen, current: currentYear),
            ),
            const SizedBox(height: AppSpacing.xl),
            if (query != null)
              _HistoryContent(query: query)
            else if (taxpayer.isLoading)
              const HistorySkeleton()
            else
              HistoryMessage(message: l10n.historyNoTaxpayer),
          ],
        ),
      ),
    );
  }

  static Future<void> _refresh(WidgetRef ref, HistoryQuery? query) async {
    if (query == null) return;
    try {
      final reloaded = ref.refresh(historyYearProvider(query).future);
      await reloaded;
    } on ApiFailure {
      // The provider holds the failure; the screen shows it with a retry.
    }
  }
}

class _HistoryContent extends ConsumerWidget {
  const new({required this.query});

  final HistoryQuery query;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final history = ref.watch(historyYearProvider(query));
    // Also called on return from a month, when this row may be gone.
    void reload() {
      if (context.mounted) ref.invalidate(historyYearProvider(query));
    }

    if (history.value case final loaded?) {
      return HistoryYearView(year: loaded, onReturn: reload);
    }
    if (history.error case final error? when !history.isLoading) {
      return HistoryFailure(error: error, onRetry: reload);
    }
    return const HistorySkeleton();
  }
}
