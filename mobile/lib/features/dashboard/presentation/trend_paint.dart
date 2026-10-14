import 'dart:math' as math;

import 'package:control_contable/core/theme/app_colors.dart';
import 'package:control_contable/features/dashboard/domain/dashboard_trend.dart';
import 'package:flutter/material.dart';

/// Surface gap between stacked segments and between paired bars (dataviz
/// mark spec), so adjacent fills never touch.
const double _gap = 2;

/// Width of a stacked bar and of each bar of a pair.
const double _stackWidth = 20;
const double _pairWidth = 14;

/// A non-zero value is never drawn thinner than this.
const double _minVisible = 2;

/// Hatch spacing and stroke on the ISR segment (secondary encoding: the
/// owner-decided IVA/ISR pair is below the normal-vision color floor).
const double _hatchStep = 6;
const double _hatchStroke = 1.5;

/// The fill of [series] in [colors].
Color seriesColor(AppColors colors, TrendSeries series) => switch (series) {
  TrendSeries.iva => colors.iva,
  TrendSeries.isr => colors.isr,
  TrendSeries.withholdings => colors.withholdings,
  TrendSeries.income => colors.income,
  TrendSeries.expense => colors.expense,
};

/// Whether [series] carries the 45° hatch.
bool isHatched(TrendSeries series) => series == TrendSeries.isr;

/// Fills [rect] with [color], hatched in [surface] when [hatched].
void paintFill(
  Canvas canvas,
  Rect rect, {
  required Color color,
  required Color surface,
  required bool hatched,
}) {
  canvas.drawRect(rect, Paint()..color = color);
  if (!hatched) return;
  final stroke = Paint()
    ..color = surface
    ..strokeWidth = _hatchStroke;
  canvas
    ..save()
    ..clipRect(rect);
  for (var x = rect.left - rect.height; x < rect.right; x += _hatchStep) {
    canvas.drawLine(
      Offset(x, rect.bottom),
      Offset(x + rect.height, rect.top),
      stroke,
    );
  }
  canvas.restore();
}

/// One month's bars: a square-topped stack (taxes, bottom up) or side by
/// side bars (income/expense), scaled to [scaleMax], on a baseline.
class MonthBarsPainter extends CustomPainter {
  /// Paints [values] of [series] in [colors].
  new({
    required this.series,
    required this.values,
    required this.scaleMax,
    required this.stacked,
    required this.colors,
  });

  /// The series drawn, in order.
  final List<TrendSeries> series;

  /// Centavos per series, same order.
  final List<int> values;

  /// The value at the top of the plot.
  final int scaleMax;

  /// Stack (taxes) or pair (income/expense).
  final bool stacked;

  /// The palette.
  final AppColors colors;

  @override
  void paint(Canvas canvas, Size size) {
    canvas.drawLine(
      Offset(0, size.height - 0.5),
      Offset(size.width, size.height - 0.5),
      Paint()..color = colors.outline,
    );
    if (scaleMax <= 0) return;
    if (stacked) {
      _paintStack(canvas, size);
    } else {
      _paintPair(canvas, size);
    }
  }

  double _heightOf(int value, Size size) => value <= 0
      ? 0
      : math.max(_minVisible, value / scaleMax * (size.height - 1));

  void _paintStack(Canvas canvas, Size size) {
    final left = (size.width - _stackWidth) / 2;
    var bottom = size.height - 1;
    for (var i = 0; i < series.length; i++) {
      final height = _heightOf(values[i], size);
      if (height == 0) continue;
      final rect = Rect.fromLTRB(
        left,
        bottom - height,
        left + _stackWidth,
        bottom,
      );
      paintFill(
        canvas,
        rect,
        color: seriesColor(colors, series[i]),
        surface: colors.surface,
        hatched: isHatched(series[i]),
      );
      bottom -= height + _gap;
    }
  }

  void _paintPair(Canvas canvas, Size size) {
    final total = series.length * _pairWidth + (series.length - 1) * _gap;
    var left = (size.width - total) / 2;
    for (var i = 0; i < series.length; i++) {
      final height = _heightOf(values[i], size);
      final bottom = size.height - 1;
      if (height > 0) {
        paintFill(
          canvas,
          Rect.fromLTRB(left, bottom - height, left + _pairWidth, bottom),
          color: seriesColor(colors, series[i]),
          surface: colors.surface,
          hatched: isHatched(series[i]),
        );
      }
      left += _pairWidth + _gap;
    }
  }

  @override
  bool shouldRepaint(MonthBarsPainter old) =>
      old.scaleMax != scaleMax ||
      old.stacked != stacked ||
      old.colors != colors ||
      !_sameList(old.values, values) ||
      !_sameList(old.series, series);

  static bool _sameList<T>(List<T> a, List<T> b) {
    if (a.length != b.length) return false;
    for (var i = 0; i < a.length; i++) {
      if (a[i] != b[i]) return false;
    }
    return true;
  }
}

/// A key swatch: the series fill (hatch included) in a small square.
class SwatchPainter extends CustomPainter {
  /// Paints [series]'s swatch.
  new({required this.series, required this.colors});

  /// The series.
  final TrendSeries series;

  /// The palette.
  final AppColors colors;

  @override
  void paint(Canvas canvas, Size size) => paintFill(
    canvas,
    Offset.zero & size,
    color: seriesColor(colors, series),
    surface: colors.surface,
    hatched: isHatched(series),
  );

  @override
  bool shouldRepaint(SwatchPainter old) =>
      old.series != series || old.colors != colors;
}
