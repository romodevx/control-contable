import 'dart:async';

import 'package:control_contable/api/generated/models/disclaimer.dart';
import 'package:control_contable/api/generated/models/disclaimer_severity.dart';
import 'package:control_contable/api/generated/models/history_year.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/estimate_legend.dart';
import 'package:control_contable/features/history/presentation/history_month_labels.dart';
import 'package:control_contable/features/history/presentation/history_month_tile.dart';
import 'package:control_contable/features/history/presentation/history_states.dart';
import 'package:control_contable/features/history/presentation/history_year_totals.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_notices.dart';
import 'package:control_contable/features/tax_engine/tax_engine_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';
import 'package:go_router/go_router.dart';

/// A loaded fiscal year, in reading order (MODULE-history §6): the strong
/// CPA banner, the year totals, the mandatory legend, the warnings, every
/// month (newest first), then the informational disclaimers.
///
/// Coming back from a month, the year is read again and the month's new
/// state ("Visto …") is announced politely once it arrives (§6).
class HistoryYearView extends StatefulWidget {
  /// Shows [year]; [onReturn] reads it again after a month or one of its
  /// record lists was opened.
  const new({required this.year, required this.onReturn, super.key});

  /// The year shown.
  final HistoryYear year;

  /// Called when the user comes back from a month or a record list.
  final VoidCallback onReturn;

  @override
  State<HistoryYearView> createState() => _HistoryYearViewState();
}

class _HistoryYearViewState extends State<HistoryYearView> {
  /// The month just returned from, announced when its re-read arrives.
  String? _returnedFrom;

  HistoryYear get year => widget.year;

  @override
  void didUpdateWidget(HistoryYearView oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (!identical(oldWidget.year, widget.year)) _announceReturn();
  }

  /// Announces the state of the month returned from, once.
  void _announceReturn() {
    final period = _returnedFrom;
    _returnedFrom = null;
    if (period == null) return;
    final month = year.months.where((m) => m.period == period).firstOrNull;
    if (month == null) return;
    final state = HistoryMonthLabels(context, month).returnState;
    if (state == null) return;
    final view = View.of(context);
    final direction = Directionality.of(context);
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      unawaited(
        // Polite by default: it never interrupts.
        SemanticsService.sendAnnouncement(view, state, direction),
      );
    });
  }

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
                onOpen: () => _open(
                  TaxEngineRoutes.estimateOf(month.period),
                  announcing: month.period,
                ),
                onOpenRecords: _open,
              ),
          ],
        ),
        for (final disclaimer in notes) DisclaimerNote(disclaimer: disclaimer),
      ],
    );
  }

  Iterable<Disclaimer> _bySeverity(DisclaimerSeverity severity) =>
      year.disclaimers.where((d) => d.severity == severity);

  /// Opens [location] (a month's estimate, M5-25, whose regimes lead to
  /// the breakdown, M5-26; or one of its record lists), then reads the
  /// year again (US-H2) and, for a month, announces its new state.
  void _open(String location, {String? announcing}) => unawaited(
    context.push<void>(location).whenComplete(() {
      if (!mounted) return;
      _returnedFrom = announcing;
      widget.onReturn();
    }),
  );
}
