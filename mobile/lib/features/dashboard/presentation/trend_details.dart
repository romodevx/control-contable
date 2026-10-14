import 'package:control_contable/api/generated/models/taxpayer_estimate_status.dart';
import 'package:control_contable/api/generated/models/trend_point.dart';
import 'package:control_contable/api/generated/models/trend_tax_iva_outcome.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/dashboard/domain/dashboard_trend.dart';
import 'package:control_contable/features/dashboard/presentation/trend_labels.dart';
import 'package:control_contable/features/tax_engine/presentation/money_row.dart';
import 'package:control_contable/features/tax_engine/presentation/status_chip.dart';
import 'package:flutter/material.dart';

/// One month's figures as text: its name, its status and one row per
/// series (plus the tax total). Used under the chart for the selected
/// month and once per month in the table view.
class TrendMonthFigures extends StatelessWidget {
  /// The figures of [point] for [series].
  const new({
    required this.point,
    required this.series,
    required this.showTaxTotal,
    super.key,
  });

  /// The month.
  final TrendPoint point;

  /// The rows shown, in order.
  final List<TrendSeries> series;

  /// Whether a "Total" row closes the tax figures.
  final bool showTaxTotal;

  @override
  Widget build(BuildContext context) {
    final labels = TrendLabels(context);
    final text = Theme.of(context).textTheme;
    final status = _statusWord(labels, point.tax?.status);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Wrap(
          spacing: AppSpacing.xs,
          runSpacing: AppSpacing.xxs,
          crossAxisAlignment: WrapCrossAlignment.center,
          children: [
            Semantics(
              header: true,
              child: Text(
                labels.monthYear(point.period),
                style: text.titleSmall,
              ),
            ),
            if (status != null) StatusChip(label: status),
          ],
        ),
        for (final s in series) _row(context, labels, s),
        if (showTaxTotal && point.tax != null)
          MoneyRow(
            label: labels.l10n.dashboardTotal,
            centavos: point.tax!.totalPayableRoundedCents,
          ),
      ],
    );
  }

  Widget _row(BuildContext context, TrendLabels labels, TrendSeries s) {
    if (s == TrendSeries.iva &&
        point.tax?.ivaOutcome == TrendTaxIvaOutcome.inFavor) {
      return Padding(
        padding: const EdgeInsets.symmetric(vertical: AppSpacing.xxs),
        child: Text(
          labels.l10n.dashboardIvaInFavor,
          style: Theme.of(context).textTheme.bodyLarge
              ?.copyWith(color: context.tokens.colors.favor),
        ),
      );
    }
    return MoneyRow(label: labels.seriesName(s), centavos: valueOf(point, s));
  }

  static String? _statusWord(
    TrendLabels labels,
    TaxpayerEstimateStatus? status,
  ) => switch (status) {
    TaxpayerEstimateStatus.complete => labels.l10n.estimateComplete,
    TaxpayerEstimateStatus.incomplete => labels.l10n.estimateIncomplete,
    TaxpayerEstimateStatus.partial => labels.l10n.estimatePartial,
    TaxpayerEstimateStatus.informational => labels.l10n.estimateInformational,
    _ => null,
  };
}

/// The data-table alternative to the chart: every month of [trend], oldest
/// first, with every figure the response holds (income, expense, IVA, ISR,
/// withholdings, total, status). Wraps under any text size.
class TrendTable extends StatelessWidget {
  /// The table of [trend].
  const new({required this.trend, super.key});

  /// The window.
  final List<TrendPoint> trend;

  @override
  Widget build(BuildContext context) {
    final series = [
      ...seriesOf(TrendView.incomeExpense, trend),
      if (trend.any((p) => p.tax != null)) ...[
        TrendSeries.iva,
        TrendSeries.isr,
        TrendSeries.withholdings,
      ],
    ];
    final hasTax = trend.any((p) => p.tax != null);
    return Column(
      key: const ValueKey('trend-table'),
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        for (final point in trend)
          Padding(
            padding: const EdgeInsets.only(bottom: AppSpacing.md),
            child: Semantics(
              container: true,
              child: TrendMonthFigures(
                point: point,
                series: series,
                showTaxTotal: hasTax,
              ),
            ),
          ),
      ],
    );
  }
}
