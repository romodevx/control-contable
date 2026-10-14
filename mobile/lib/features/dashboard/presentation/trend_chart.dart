import 'package:control_contable/api/generated/models/trend_point.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/dashboard/domain/dashboard_trend.dart';
import 'package:control_contable/features/dashboard/presentation/trend_labels.dart';
import 'package:control_contable/features/dashboard/presentation/trend_paint.dart';
import 'package:flutter/material.dart';

/// Width of one month column: a ≥ 44 px target, so a full year scrolls
/// sideways (the current month, at the end, starts in view).
const double _columnWidth = 48;

/// Plot height; fixed, so 200 % text never squeezes the bars.
const double _plotHeight = 160;

const double _markerHeight = 20;
const double _swatchSize = 12;

/// The bars of [view] over [trend], one tappable column per month, with
/// [selected] highlighted. The whole chart is summarized for screen
/// readers by its caller; each column announces its own month.
class TrendChart extends StatelessWidget {
  /// The chart of [view].
  const new({
    required this.view,
    required this.trend,
    required this.selected,
    required this.onSelected,
    super.key,
  });

  /// Taxes or income/expense.
  final TrendView view;

  /// The window, oldest first.
  final List<TrendPoint> trend;

  /// The selected month (`YYYY-MM`).
  final String selected;

  /// A month was tapped.
  final ValueChanged<String> onSelected;

  @override
  Widget build(BuildContext context) {
    final series = seriesOf(view, trend);
    final scaleMax = scaleMaxOf(view, trend);
    final labels = TrendLabels(context);
    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      reverse: true,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          for (final point in trend)
            _MonthColumn(
              key: ValueKey('trend-bar-${point.period}'),
              point: point,
              series: series,
              view: view,
              scaleMax: scaleMax,
              isSelected: point.period == selected,
              label: labels.barSentence(view, point, series),
              short: labels.monthShort(point.period),
              onTap: () => onSelected(point.period),
            ),
        ],
      ),
    );
  }
}

class _MonthColumn extends StatelessWidget {
  const new({
    required this.point,
    required this.series,
    required this.view,
    required this.scaleMax,
    required this.isSelected,
    required this.label,
    required this.short,
    required this.onTap,
    super.key,
  });

  final TrendPoint point;
  final List<TrendSeries> series;
  final TrendView view;
  final int scaleMax;
  final bool isSelected;
  final String label;
  final String short;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final tokens = context.tokens;
    final colors = tokens.colors;
    final text = Theme.of(context).textTheme;
    final flagged = view == TrendView.taxes && isFlagged(point);
    // excludeSemantics drops the GestureDetector's own tap action, so the
    // node carries onTap itself: a "button" a screen reader can activate
    // (QA M6-06 D-1).
    return Semantics(
      button: true,
      selected: isSelected,
      label: label,
      onTap: onTap,
      excludeSemantics: true,
      child: GestureDetector(
        behavior: HitTestBehavior.opaque,
        onTap: onTap,
        child: DecoratedBox(
          decoration: BoxDecoration(
            color: isSelected ? colors.border : null,
            borderRadius: BorderRadius.circular(tokens.radii.control),
          ),
          child: SizedBox(
            width: _columnWidth,
            child: Column(
              children: [
                SizedBox(
                  height: _markerHeight,
                  child: flagged
                      ? FittedBox(
                          child: Text(
                            '▲',
                            style: text.bodySmall?.copyWith(color: colors.text),
                          ),
                        )
                      : null,
                ),
                SizedBox(
                  height: _plotHeight,
                  width: _columnWidth,
                  child: CustomPaint(
                    painter: MonthBarsPainter(
                      series: series,
                      values: [for (final s in series) valueOf(point, s)],
                      scaleMax: scaleMax,
                      stacked: view == TrendView.taxes,
                      colors: colors,
                    ),
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.symmetric(vertical: AppSpacing.xxs),
                  child: FittedBox(
                    fit: BoxFit.scaleDown,
                    child: Text(
                      short,
                      style: text.bodySmall?.copyWith(
                        color: colors.text,
                        fontWeight: isSelected ? FontWeight.w700 : null,
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

/// The chart's key: a swatch and a word per series, plus the marker's
/// meaning on the tax view. Text in text ink, never in the series color.
class TrendKey extends StatelessWidget {
  /// The key of [series].
  const new({required this.series, required this.showMarker, super.key});

  /// The series drawn.
  final List<TrendSeries> series;

  /// Whether the "incomplete or partial" marker is explained.
  final bool showMarker;

  @override
  Widget build(BuildContext context) {
    final colors = context.tokens.colors;
    final labels = TrendLabels(context);
    final style = Theme.of(context).textTheme.bodyMedium
        ?.copyWith(color: colors.text);
    return Wrap(
      spacing: AppSpacing.md,
      runSpacing: AppSpacing.xs,
      children: [
        for (final s in series)
          Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              CustomPaint(
                size: const Size.square(_swatchSize),
                painter: SwatchPainter(series: s, colors: colors),
              ),
              const SizedBox(width: AppSpacing.xs),
              Text(labels.seriesName(s), style: style),
            ],
          ),
        if (showMarker) Text(labels.l10n.dashboardMarkerKey, style: style),
      ],
    );
  }
}
