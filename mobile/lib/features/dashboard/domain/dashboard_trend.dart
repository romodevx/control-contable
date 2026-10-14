import 'package:control_contable/api/generated/models/dashboard.dart';
import 'package:control_contable/api/generated/models/taxpayer_estimate_status.dart';
import 'package:control_contable/api/generated/models/trend_point.dart';

/// The two views of the Home trend chart (MODULE-dashboard US-D3).
enum TrendView {
  /// Stacked estimated taxes per month (the default).
  taxes,

  /// Income collected and expenses paid, side by side per month.
  incomeExpense,
}

/// One series of the trend chart, in drawing order.
enum TrendSeries {
  /// IVA to pay (bottom of the tax stack).
  iva,

  /// ISR to pay (on top of IVA).
  isr,

  /// Withholdings to remit (top of the stack, only when any month has some).
  withholdings,

  /// "Ingresos cobrados" (before IVA).
  income,

  /// "Gastos pagados" (before IVA).
  expense,
}

/// The views [trend] can show: taxes when the engine is ON (any `tax`),
/// income/expense when either record module is ON (any non-null figure).
/// The server nulls what a disabled module would show (§3), so the
/// response alone decides.
List<TrendView> trendViewsOf(List<TrendPoint> trend) => [
  if (trend.any((p) => p.tax != null)) TrendView.taxes,
  if (trend.any((p) => p.incomeCents != null || p.expenseCents != null))
    TrendView.incomeExpense,
];

/// The series [view] draws for [trend], bottom to top (taxes) or left to
/// right (income/expense).
List<TrendSeries> seriesOf(TrendView view, List<TrendPoint> trend) =>
    switch (view) {
      TrendView.taxes => [
        TrendSeries.iva,
        TrendSeries.isr,
        if (trend.any((p) => (p.tax?.withholdingsToRemitCents ?? 0) > 0))
          TrendSeries.withholdings,
      ],
      TrendView.incomeExpense => [
        if (trend.any((p) => p.incomeCents != null)) TrendSeries.income,
        if (trend.any((p) => p.expenseCents != null)) TrendSeries.expense,
      ],
    };

/// [series] of [point] in centavos; zero when its module is OFF.
int valueOf(TrendPoint point, TrendSeries series) => switch (series) {
  TrendSeries.iva => point.tax?.ivaPayableRoundedCents ?? 0,
  TrendSeries.isr => point.tax?.isrPayableRoundedCents ?? 0,
  TrendSeries.withholdings => point.tax?.withholdingsToRemitCents ?? 0,
  TrendSeries.income => point.incomeCents ?? 0,
  TrendSeries.expense => point.expenseCents ?? 0,
};

/// The top of the value axis for [view]: the tallest stack (taxes) or the
/// tallest single bar (income/expense). One pass over [trend].
int scaleMaxOf(TrendView view, List<TrendPoint> trend) {
  final series = seriesOf(view, trend);
  var max = 0;
  for (final point in trend) {
    final values = series.map((s) => valueOf(point, s));
    final height = view == TrendView.taxes
        ? values.fold(0, (sum, v) => sum + v)
        : values.fold(0, (top, v) => v > top ? v : top);
    if (height > max) max = height;
  }
  return max;
}

/// The month with the highest estimated total (the first on a tie), or
/// `null` when every month is zero or the engine is OFF.
TrendPoint? peakTaxMonth(List<TrendPoint> trend) {
  TrendPoint? peak;
  var peakTotal = 0;
  for (final point in trend) {
    final total = point.tax?.totalPayableRoundedCents ?? 0;
    if (total > peakTotal) {
      peak = point;
      peakTotal = total;
    }
  }
  return peak;
}

/// [series] summed over the whole window.
int windowTotalOf(TrendSeries series, List<TrendPoint> trend) =>
    trend.fold(0, (sum, point) => sum + valueOf(point, series));

/// Whether [point]'s figures depend on missing inputs or cover only some
/// taxes (its bar carries a marker; the text says which).
bool isFlagged(TrendPoint point) => switch (point.tax?.status) {
  TaxpayerEstimateStatus.incomplete || TaxpayerEstimateStatus.partial => true,
  _ => false,
};

/// The taxpayer-level status of the Home hero: the last trend point's
/// (contract: "the hero chip is the status of the last point"); `none`
/// when the engine is OFF.
TaxpayerEstimateStatus heroStatusOf(Dashboard dashboard) =>
    dashboard.trend.lastOrNull?.tax?.status ?? TaxpayerEstimateStatus.none;
