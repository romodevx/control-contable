import 'package:control_contable/api/generated/models/taxpayer_estimate_status.dart';
import 'package:control_contable/api/generated/models/trend_point.dart';
import 'package:control_contable/features/dashboard/domain/dashboard_trend.dart';
import 'package:flutter_test/flutter_test.dart';

import 'dashboard_fixtures.dart';

List<TrendPoint> _trend(List<Map<String, Object?>> json) =>
    dashboardFrom(dashboardJson(trend: json)).trend;

void main() {
  group('trendViewsOf', () {
    test('taxes first (the default), then income and expense', () {
      expect(trendViewsOf(_trend(threeMonthTrend())), [
        TrendView.taxes,
        TrendView.incomeExpense,
      ]);
    });

    test('engine OFF: only income and expense', () {
      final trend = dashboardFrom(dashboardJson(engine: false)).trend;
      expect(trendViewsOf(trend), [TrendView.incomeExpense]);
    });

    test('both record modules OFF: only taxes', () {
      expect(
        trendViewsOf(
          _trend([trendPointJson('2026-03', income: null, expense: null)]),
        ),
        [TrendView.taxes],
      );
    });

    test('nothing at all: no view', () {
      final trend = dashboardFrom(
        dashboardJson(
          engine: false,
          trend: [
            trendPointJson('2026-03', income: null, expense: null, tax: null),
          ],
        ),
      ).trend;
      expect(trendViewsOf(trend), isEmpty);
    });
  });

  group('seriesOf', () {
    test('taxes: IVA under ISR; withholdings only when a month has some', () {
      expect(seriesOf(TrendView.taxes, _trend(threeMonthTrend())), [
        TrendSeries.iva,
        TrendSeries.isr,
      ]);
      final withHeld = _trend([
        trendPointJson('2026-03', tax: trendTaxJson(iva: 1, withholdings: 5)),
      ]);
      expect(seriesOf(TrendView.taxes, withHeld), [
        TrendSeries.iva,
        TrendSeries.isr,
        TrendSeries.withholdings,
      ]);
    });

    test('income and expense: a module OFF drops its series', () {
      final noIncome = _trend([trendPointJson('2026-03', income: null)]);
      expect(seriesOf(TrendView.incomeExpense, noIncome), [
        TrendSeries.expense,
      ]);
      final noExpense = _trend([trendPointJson('2026-03', expense: null)]);
      expect(seriesOf(TrendView.incomeExpense, noExpense), [
        TrendSeries.income,
      ]);
    });
  });

  test('valueOf reads centavos, zero when absent', () {
    final point = _trend([
      trendPointJson(
        '2026-03',
        income: 7,
        expense: null,
        tax: trendTaxJson(iva: 1, isr: 2, withholdings: 3),
      ),
    ]).single;
    expect(valueOf(point, TrendSeries.iva), 1);
    expect(valueOf(point, TrendSeries.isr), 2);
    expect(valueOf(point, TrendSeries.withholdings), 3);
    expect(valueOf(point, TrendSeries.income), 7);
    expect(valueOf(point, TrendSeries.expense), 0);
    final off = _trend([trendPointJson('2026-03', tax: null)]).single;
    expect(valueOf(off, TrendSeries.iva), 0);
  });

  group('scaleMaxOf', () {
    test('taxes stack: the tallest month sum', () {
      expect(scaleMaxOf(TrendView.taxes, _trend(threeMonthTrend())), 350000);
    });

    test('income and expense sit side by side: the tallest bar', () {
      expect(
        scaleMaxOf(TrendView.incomeExpense, _trend(threeMonthTrend())),
        3000000,
      );
    });
  });

  group('peakTaxMonth', () {
    test('the month with the highest total; the first on a tie', () {
      expect(peakTaxMonth(_trend(threeMonthTrend()))?.period, '2026-02');
      final tie = _trend([
        trendPointJson('2026-01', tax: trendTaxJson(iva: 5)),
        trendPointJson('2026-02', tax: trendTaxJson(isr: 5)),
      ]);
      expect(peakTaxMonth(tie)?.period, '2026-01');
    });

    test('none when every month is zero or the engine is OFF', () {
      expect(peakTaxMonth(_trend([trendPointJson('2026-01')])), isNull);
      expect(
        peakTaxMonth(_trend([trendPointJson('2026-01', tax: null)])),
        isNull,
      );
    });
  });

  test('windowTotalOf adds a series over the window', () {
    final trend = _trend(threeMonthTrend());
    expect(windowTotalOf(TrendSeries.income, trend), 6500000);
    expect(windowTotalOf(TrendSeries.expense, trend), 2000000);
  });

  test('isFlagged marks incomplete and partial months only', () {
    TrendPoint withStatus(String status) => _trend([
      trendPointJson('2026-03', tax: trendTaxJson(status: status)),
    ]).single;
    expect(isFlagged(withStatus('incomplete')), isTrue);
    expect(isFlagged(withStatus('partial')), isTrue);
    expect(isFlagged(withStatus('complete')), isFalse);
    expect(isFlagged(withStatus('none')), isFalse);
    expect(isFlagged(_trend([trendPointJson('2026-03', tax: null)]).single),
        isFalse);
  });

  test('heroStatusOf is the status of the last point', () {
    expect(
      heroStatusOf(dashboardFrom(dashboardJson())),
      TaxpayerEstimateStatus.complete,
    );
    expect(
      heroStatusOf(dashboardFrom(dashboardJson(engine: false))),
      TaxpayerEstimateStatus.none,
    );
  });
}
