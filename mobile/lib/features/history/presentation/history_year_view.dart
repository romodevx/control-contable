import 'dart:async';

import 'package:control_contable/api/generated/models/disclaimer.dart';
import 'package:control_contable/api/generated/models/disclaimer_severity.dart';
import 'package:control_contable/api/generated/models/history_year.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/estimate_legend.dart';
import 'package:control_contable/features/history/presentation/history_month_tile.dart';
import 'package:control_contable/features/history/presentation/history_states.dart';
import 'package:control_contable/features/history/presentation/history_year_totals.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_notices.dart';
import 'package:control_contable/features/tax_engine/tax_engine_routes.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// A loaded fiscal year, in reading order (MODULE-history §6): the strong
/// CPA banner, the year totals, the mandatory legend, the warnings, every
/// month (newest first), then the informational disclaimers.
class HistoryYearView extends StatelessWidget {
  /// Shows [year]; [onReturn] reads it again after a month was opened.
  const new({required this.year, required this.onReturn, super.key});

  /// The year shown.
  final HistoryYear year;

  /// Called when the user comes back from a month.
  final VoidCallback onReturn;

  @override
  Widget build(BuildContext context) {
    if (year.months.every((month) => !month.hasRecords)) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        spacing: AppSpacing.md,
        children: [
          const HistoryEmpty(),
          EstimateLegend(text: year.legend),
        ],
      );
    }
    final strong = _bySeverity(DisclaimerSeverity.strong);
    final warnings = _bySeverity(DisclaimerSeverity.warning);
    final notes = year.disclaimers.where(
      (d) =>
          d.severity == DisclaimerSeverity.info ||
          d.severity == DisclaimerSeverity.$unknown,
    );
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      spacing: AppSpacing.md,
      children: [
        for (final (index, disclaimer) in strong.indexed)
          CpaReviewBanner(
            key: index == 0 ? const ValueKey('history-cpa-banner') : null,
            disclaimer: disclaimer,
          ),
        HistoryYearTotalsCard(totals: year.yearTotals),
        EstimateLegend(text: year.legend),
        for (final disclaimer in warnings)
          DisclaimerWarning(disclaimer: disclaimer),
        Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            for (final month in year.months)
              HistoryMonthTile(
                key: ValueKey('history-month-${month.period}'),
                month: month,
                onOpen: () => _open(context, month.period),
              ),
          ],
        ),
        for (final disclaimer in notes) DisclaimerNote(disclaimer: disclaimer),
      ],
    );
  }

  Iterable<Disclaimer> _bySeverity(DisclaimerSeverity severity) =>
      year.disclaimers.where((d) => d.severity == severity);

  /// Opens the month's estimate (M5-25; its regimes lead to the breakdown,
  /// M5-26), then reads the year again (US-H2).
  void _open(BuildContext context, String period) => unawaited(
    context
        .push<void>(TaxEngineRoutes.estimateOf(period))
        .whenComplete(onReturn),
  );
}
