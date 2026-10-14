import 'package:control_contable/api/generated/models/taxpayer_estimate_status.dart';
import 'package:control_contable/api/generated/models/trend_point.dart';
import 'package:control_contable/api/generated/models/trend_tax_iva_outcome.dart';
import 'package:control_contable/features/dashboard/domain/dashboard_trend.dart';
import 'package:control_contable/features/tax_engine/domain/money_format.dart';
import 'package:control_contable/features/taxpayers/domain/periods.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

/// Words and sentences of the trend chart, its key and its table, built
/// from one response so the chart, the text and the screen reader always
/// say the same thing.
class TrendLabels {
  /// Labels in [context]'s locale.
  new(BuildContext context)
    : l10n = AppLocalizations.of(context),
      _material = MaterialLocalizations.of(context),
      _locale = Localizations.localeOf(context).toString();

  /// The app's copy.
  final AppLocalizations l10n;
  final MaterialLocalizations _material;
  final String _locale;

  /// "marzo de 2026".
  String monthYear(String period) =>
      _material.formatMonthYear(periodMonth(period));

  /// "marzo".
  String monthName(String period) =>
      DateFormat.MMMM(_locale).format(periodMonth(period));

  /// "mar", under a bar.
  String monthShort(String period) =>
      DateFormat.MMM(_locale).format(periodMonth(period)).replaceAll('.', '');

  /// The key word of [series].
  String seriesName(TrendSeries series) => switch (series) {
    TrendSeries.iva => l10n.dashboardSeriesIva,
    TrendSeries.isr => l10n.dashboardSeriesIsr,
    TrendSeries.withholdings => l10n.dashboardSeriesWithholdings,
    TrendSeries.income => l10n.dashboardIncomeCollected,
    TrendSeries.expense => l10n.dashboardExpensePaid,
  };

  /// The status word of a flagged month ("Incompleto", "Parcial"), or
  /// `null` when its figures are final or there are none.
  String? statusWord(TrendPoint point) => switch (point.tax?.status) {
    TaxpayerEstimateStatus.incomplete => l10n.estimateIncomplete,
    TaxpayerEstimateStatus.partial => l10n.estimatePartial,
    _ => null,
  };

  /// What one month's bar announces: its month, each series with its
  /// amount, the tax total and the status.
  String barSentence(TrendView view, TrendPoint point, List<TrendSeries> all) {
    final total = point.tax?.totalPayableRoundedCents;
    final totalWord = l10n.dashboardTotal.toLowerCase();
    final parts = [
      for (final series in all) _spokenSeries(point, series),
      if (view == TrendView.taxes && total != null)
        '$totalWord ${spokenMoney(total)}',
      if (view == TrendView.taxes && statusWord(point) != null)
        statusWord(point)!.toLowerCase(),
    ];
    return '${monthYear(point.period)}: ${parts.join(', ')}';
  }

  String _spokenSeries(TrendPoint point, TrendSeries series) {
    if (series == TrendSeries.iva &&
        point.tax?.ivaOutcome == TrendTaxIvaOutcome.inFavor) {
      return l10n.dashboardIvaInFavor;
    }
    final name = seriesName(series);
    final spokenName =
        series == TrendSeries.income || series == TrendSeries.expense
        ? name.toLowerCase()
        : name;
    return '$spokenName ${spokenMoney(valueOf(point, series))}';
  }

  /// The chart's screen-reader summary (MODULE-dashboard §6).
  String summary(TrendView view, List<TrendPoint> trend) {
    final current = trend.last;
    final month = _capitalized(monthName(current.period));
    if (view == TrendView.taxes) {
      final status = statusWord(current);
      final peak = peakTaxMonth(trend);
      final head = l10n.dashboardTaxSummary(
        trend.length,
        month,
        spokenMoney(current.tax?.totalPayableRoundedCents ?? 0),
        status == null ? '' : ', ${status.toLowerCase()}',
      );
      if (peak == null) return head;
      return head +
          l10n.dashboardTaxPeak(
            monthName(peak.period),
            spokenMoney(peak.tax!.totalPayableRoundedCents),
          );
    }
    final series = seriesOf(view, trend);
    String spoken(int Function(TrendSeries) amountOf) => [
      for (final s in series)
        '${seriesName(s).toLowerCase()} ${spokenMoney(amountOf(s))}',
    ].join(', ');
    return l10n.dashboardIncomeExpenseSummary(
      trend.length,
      month,
      spoken((s) => valueOf(current, s)),
      spoken((s) => windowTotalOf(s, trend)),
    );
  }

  static String _capitalized(String word) =>
      word.isEmpty ? word : word[0].toUpperCase() + word.substring(1);
}
