import 'package:control_contable/api/generated/models/trend_point.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/estimate_legend.dart';
import 'package:control_contable/features/dashboard/domain/dashboard_trend.dart';
import 'package:control_contable/features/dashboard/presentation/trend_chart.dart';
import 'package:control_contable/features/dashboard/presentation/trend_details.dart';
import 'package:control_contable/features/dashboard/presentation/trend_labels.dart';
import 'package:flutter/material.dart';

/// The 12-month trend (MODULE-dashboard US-D3): a view switcher when both
/// views exist, the chart with its key and the selected month's values as
/// text (or the table instead), and the mandatory legend under any tax
/// figure. One chart, no animation.
class TrendSection extends StatefulWidget {
  /// The section over [trend]; [legend] is the server's, present iff the
  /// engine is ON.
  const new({required this.trend, required this.legend, super.key});

  /// The window, oldest first (never empty).
  final List<TrendPoint> trend;

  /// The server's mandatory legend.
  final String? legend;

  @override
  State<TrendSection> createState() => _TrendSectionState();
}

class _TrendSectionState extends State<TrendSection> {
  TrendView? _view;
  String? _selected;
  bool _asTable = false;

  @override
  Widget build(BuildContext context) {
    final trend = widget.trend;
    final views = trendViewsOf(trend);
    final view = views.contains(_view) ? _view! : views.first;
    final selected = trend.any((p) => p.period == _selected)
        ? _selected!
        : trend.last.period;
    final labels = TrendLabels(context);
    final l10n = labels.l10n;
    final showsTax = _asTable
        ? trend.any((p) => p.tax != null)
        : view == TrendView.taxes;
    final series = seriesOf(view, trend);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Semantics(
          header: true,
          child: Text(
            l10n.dashboardTrendTitle(trend.length),
            style: Theme.of(context).textTheme.titleMedium,
          ),
        ),
        const SizedBox(height: AppSpacing.sm),
        if (views.length > 1 && !_asTable) ...[
          SegmentedButton<TrendView>(
            showSelectedIcon: false,
            segments: [
              for (final v in views)
                ButtonSegment(
                  value: v,
                  label: Text(
                    v == TrendView.taxes
                        ? l10n.dashboardViewTaxes
                        : l10n.dashboardViewIncomeExpense,
                  ),
                ),
            ],
            selected: {view},
            onSelectionChanged: (choice) =>
                setState(() => _view = choice.single),
          ),
          const SizedBox(height: AppSpacing.md),
        ],
        if (_asTable)
          TrendTable(trend: trend)
        else ...[
          Semantics(
            container: true,
            label: labels.summary(view, trend),
            child: TrendChart(
              key: const ValueKey('trend-chart'),
              view: view,
              trend: trend,
              selected: selected,
              onSelected: (period) => setState(() => _selected = period),
            ),
          ),
          const SizedBox(height: AppSpacing.sm),
          TrendKey(
            key: const ValueKey('trend-key'),
            series: series,
            showMarker: view == TrendView.taxes && trend.any(isFlagged),
          ),
          const SizedBox(height: AppSpacing.md),
          TrendMonthFigures(
            key: const ValueKey('trend-selection'),
            point: trend.firstWhere((p) => p.period == selected),
            series: series,
            showTaxTotal: view == TrendView.taxes,
          ),
        ],
        if (showsTax && widget.legend != null) ...[
          const SizedBox(height: AppSpacing.sm),
          EstimateLegend(text: widget.legend),
        ],
        Align(
          alignment: AlignmentDirectional.centerStart,
          child: TextButton(
            onPressed: () => setState(() => _asTable = !_asTable),
            child: Text(
              _asTable ? l10n.dashboardShowChart : l10n.dashboardShowTable,
            ),
          ),
        ),
      ],
    );
  }
}
