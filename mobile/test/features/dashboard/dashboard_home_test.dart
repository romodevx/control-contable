import 'dart:async';

import 'package:control_contable/api/generated/models/dashboard.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/taxpayers/application/active_taxpayer.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../tax_engine/estimate_fixtures.dart';
import '../taxpayers/taxpayers_fakes.dart';
import 'dashboard_fixtures.dart';
import 'pump_dashboard.dart';

const String _placeholder = 'Aquí aparecerá el estimado del mes.';
const String _emptyCopy =
    'Aún no hay ingresos ni gastos registrados. El estimado aparece con el '
    'primer registro.';

/// A tall phone, so the whole Home is built without scrolling.
void _tall(WidgetTester tester) {
  tester.view
    ..physicalSize = const Size(400, 3000)
    ..devicePixelRatio = 1;
  addTearDown(tester.view.reset);
}

Finder _within(String key, Finder finder) =>
    find.descendant(of: find.byKey(ValueKey(key)), matching: finder);

double _top(WidgetTester tester, Finder finder) =>
    tester.getTopLeft(finder.first).dy;

Dashboard _reply({
  bool isEmpty = false,
  int? monthIncome = 3000000,
  int? monthExpense = 1000000,
  Map<String, Object?>? estimate,
  bool engine = true,
  List<Map<String, Object?>>? trend,
  Map<String, Object?>? access,
}) => dashboardFrom(
  dashboardJson(
    isEmpty: isEmpty,
    monthIncome: monthIncome,
    monthExpense: monthExpense,
    estimate: estimate,
    engine: engine,
    trend: trend,
    access: access,
  ),
);

void main() {
  late DashboardHarness harness;

  setUp(() => harness = DashboardHarness());

  group('module_dashboard', () {
    testWidgets('OFF: the previous home stays and nothing is asked', (
      tester,
    ) async {
      await harness.pump(
        tester,
        modules: homeModules.difference({AppModule.dashboard}),
      );

      expect(find.text(_placeholder), findsOneWidget);
      expect(harness.repository.calls, isEmpty);
    });

    testWidgets('ON: Home asks for the active taxpayer', (tester) async {
      _tall(tester);
      await harness.pump(tester);

      expect(find.text(_placeholder), findsNothing);
      expect(harness.repository.calls, [fixtureTaxpayerId]);
      expect(find.byKey(const ValueKey('dashboard-hero')), findsOneWidget);
    });
  });

  group("this month's estimate", () {
    testWidgets('totals, status chip and the legend right under them', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(tester);

      expect(
        _within('dashboard-hero', find.text('Total a pagar')),
        findsOneWidget,
      );
      expect(
        find.bySemanticsLabel('Total a pagar: 2,284 pesos con 29 centavos'),
        findsOneWidget,
      );
      expect(
        _within('dashboard-status', find.text('Completo')),
        findsOneWidget,
      );
      final legend = find.text(legendText).first;
      expect(
        _top(tester, legend),
        greaterThan(_top(tester, find.byKey(const ValueKey('dashboard-hero')))),
      );
      expect(
        _top(tester, legend),
        lessThan(_top(tester, find.byKey(const ValueKey('dashboard-month')))),
      );
    });

    testWidgets('"Ver detalle" opens the estimate screen', (tester) async {
      _tall(tester);
      await harness.pump(tester);

      await tester.tap(find.text('Ver detalle'));
      await tester.pumpAndSettle();

      expect(harness.location, '/estimado');
    });

    testWidgets('incomplete: "Completa tus datos" leads to the estimate', (
      tester,
    ) async {
      _tall(tester);
      final trend = threeMonthTrend()
        ..last = trendPointJson(
          '2026-03',
          tax: trendTaxJson(iva: 213333, isr: 15096, status: 'incomplete'),
        );
      await harness.pump(tester, reply: _reply(trend: trend));

      await tester.tap(find.text('Completa tus datos'));
      await tester.pumpAndSettle();

      expect(harness.location, '/estimado');
    });

    testWidgets('the strong CPA banner sits above the hero', (tester) async {
      _tall(tester);
      await harness.pump(
        tester,
        reply: _reply(
          estimate: estimateJson(
            disclaimers: [notTaxAdvice, cpaReviewDisclaimer],
          ),
        ),
      );

      final banner = find.byKey(const ValueKey('cpa-review-banner'));
      expect(banner, findsOneWidget);
      expect(
        _top(tester, banner),
        lessThan(_top(tester, find.byKey(const ValueKey('dashboard-hero')))),
      );
      expect(find.text(notTaxAdvice['text']! as String), findsNothing);
    });

    testWidgets('only informational regimes: an explanation, no amount', (
      tester,
    ) async {
      _tall(tester);
      final trend = threeMonthTrend()
        ..last = trendPointJson(
          '2026-03',
          tax: trendTaxJson(status: 'informational'),
        );
      await harness.pump(tester, reply: _reply(trend: trend));

      expect(
        find.text('Este régimen es informativo: la app no estima su impuesto'),
        findsOneWidget,
      );
      expect(find.text('Total a pagar'), findsNothing);
      expect(find.text('Ver detalle'), findsOneWidget);
    });

    testWidgets('no regime this month: the add-a-regime prompt instead', (
      tester,
    ) async {
      _tall(tester);
      final trend = threeMonthTrend()
        ..last = trendPointJson('2026-03', tax: trendTaxJson(status: 'none'));
      await harness.pump(
        tester,
        held: [
          taxpayer(regimes: [heldRegime('612', endPeriod: '2026-02')]),
        ],
        reply: _reply(trend: trend),
      );

      expect(
        find.text('Agrega un régimen para ver tu estimado.'),
        findsOneWidget,
      );
      expect(find.byKey(const ValueKey('dashboard-hero')), findsNothing);
      expect(find.byKey(const ValueKey('trend-chart')), findsOneWidget);
    });

    testWidgets('engine OFF: no tax figure, no legend, no view switcher', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(
        tester,
        modules: homeModules.difference({AppModule.taxEngine}),
        reply: _reply(engine: false),
      );

      expect(find.byKey(const ValueKey('dashboard-hero')), findsNothing);
      expect(find.text(legendText), findsNothing);
      expect(find.text('Impuestos'), findsNothing);
      expect(find.text('Ingresos cobrados'), findsWidgets);
      expect(find.byKey(const ValueKey('trend-chart')), findsOneWidget);
    });
  });

  group('income vs expense this month', () {
    testWidgets('collected, paid and their difference', (tester) async {
      _tall(tester);
      await harness.pump(tester);

      expect(
        _within('dashboard-month', find.text('Ingresos cobrados')),
        findsOneWidget,
      );
      expect(
        _within('dashboard-month', find.text(r'$30,000.00')),
        findsOneWidget,
      );
      expect(
        _within('dashboard-month', find.text('Gastos pagados')),
        findsOneWidget,
      );
      expect(
        _within('dashboard-month', find.text(r'$10,000.00')),
        findsOneWidget,
      );
      expect(
        _within('dashboard-month', find.text('Diferencia')),
        findsOneWidget,
      );
      expect(
        _within('dashboard-month', find.text(r'$20,000.00')),
        findsOneWidget,
      );
    });

    testWidgets('incomes OFF: no income figure, series or quick action', (
      tester,
    ) async {
      _tall(tester);
      final trend = [
        for (final point in threeMonthTrend()) {...point, 'income_cents': null},
      ];
      await harness.pump(
        tester,
        modules: homeModules.difference({AppModule.incomes}),
        reply: _reply(monthIncome: null, trend: trend),
      );

      expect(find.text('Ingresos cobrados'), findsNothing);
      expect(find.text('Diferencia'), findsNothing);
      expect(find.text('Registrar ingreso'), findsNothing);
      expect(find.text('Gastos pagados'), findsWidgets);
      expect(find.text('Registrar gasto'), findsOneWidget);
    });

    testWidgets('both record modules OFF: no month card', (tester) async {
      _tall(tester);
      final trend = [
        for (final point in threeMonthTrend())
          {...point, 'income_cents': null, 'expense_cents': null},
      ];
      await harness.pump(
        tester,
        modules: homeModules.difference({
          AppModule.incomes,
          AppModule.expenses,
        }),
        reply: _reply(monthIncome: null, monthExpense: null, trend: trend),
      );

      expect(find.byKey(const ValueKey('dashboard-month')), findsNothing);
      expect(find.text('Ingresos y gastos'), findsNothing);
    });
  });

  group('trend chart', () {
    testWidgets('taxes by default, with a screen-reader summary', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(tester);

      expect(find.text('Impuestos'), findsOneWidget);
      expect(find.text('Ingresos y gastos'), findsOneWidget);
      expect(
        find.bySemanticsLabel(
          'Impuestos estimados, últimos 3 meses. Marzo: 2,284 pesos con 29 '
          'centavos. Mes más alto: febrero, 3,500 pesos.',
        ),
        findsOneWidget,
      );
      expect(_within('trend-key', find.text('IVA')), findsOneWidget);
      expect(_within('trend-key', find.text('ISR')), findsOneWidget);
      expect(_within('trend-key', find.text('Retenciones')), findsNothing);
      // The hero's legend and the one under the tax chart.
      expect(find.text(legendText), findsNWidgets(2));
    });

    testWidgets('withholdings show as a third series when any month has', (
      tester,
    ) async {
      _tall(tester);
      final trend = threeMonthTrend()
        ..last = trendPointJson(
          '2026-03',
          tax: trendTaxJson(iva: 1000, isr: 1000, withholdings: 50000),
        );
      await harness.pump(tester, reply: _reply(trend: trend));

      expect(_within('trend-key', find.text('Retenciones')), findsOneWidget);
    });

    testWidgets('the current month is selected and its values shown', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(tester);

      expect(
        _within('trend-selection', find.text('marzo de 2026')),
        findsOneWidget,
      );
      expect(
        _within('trend-selection', find.text(r'$2,284.29')),
        findsOneWidget,
      );
    });

    testWidgets('tapping a month shows its values and status', (tester) async {
      _tall(tester);
      await harness.pump(tester);

      await tester.tap(find.byKey(const ValueKey('trend-bar-2026-02')));
      await tester.pumpAndSettle();

      expect(
        _within('trend-selection', find.text('febrero de 2026')),
        findsOneWidget,
      );
      expect(
        _within('trend-selection', find.text('Incompleto')),
        findsOneWidget,
      );
      expect(
        find.bySemanticsLabel(
          RegExp('^febrero de 2026: IVA 3,100 pesos.*incompleto'),
        ),
        findsOneWidget,
      );
    });

    testWidgets('a saldo a favor month says so, never a negative bar', (
      tester,
    ) async {
      _tall(tester);
      final trend = threeMonthTrend()
        ..last = trendPointJson(
          '2026-03',
          tax: trendTaxJson(isr: 15096, ivaOutcome: 'in_favor'),
        );
      await harness.pump(tester, reply: _reply(trend: trend));

      expect(
        _within('trend-selection', find.text('IVA: saldo a favor')),
        findsOneWidget,
      );
    });

    testWidgets('income and expense view', (tester) async {
      _tall(tester);
      await harness.pump(tester);

      await tester.tap(find.text('Ingresos y gastos'));
      await tester.pumpAndSettle();

      expect(
        find.bySemanticsLabel(
          'Ingresos y gastos, últimos 3 meses. Marzo: ingresos cobrados '
          '30,000 pesos, gastos pagados 10,000 pesos. En el periodo: '
          'ingresos cobrados 65,000 pesos, gastos pagados 20,000 pesos.',
        ),
        findsOneWidget,
      );
      // Only the hero's legend: this view shows no tax figure.
      expect(find.text(legendText), findsOneWidget);
    });

    testWidgets('"Ver como tabla" shows the same data as a table', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(tester);

      await tester.tap(find.text('Ver como tabla'));
      await tester.pumpAndSettle();

      expect(find.byKey(const ValueKey('trend-chart')), findsNothing);
      expect(
        _within('trend-table', find.text('enero de 2026')),
        findsOneWidget,
      );
      expect(
        _within('trend-table', find.text('febrero de 2026')),
        findsOneWidget,
      );
      expect(_within('trend-table', find.text('Incompleto')), findsOneWidget);
      expect(find.text(legendText), findsNWidgets(2));

      await tester.tap(find.text('Ver como gráfica'));
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('trend-chart')), findsOneWidget);
    });
  });

  group('access banner', () {
    testWidgets('trial: days and records left', (tester) async {
      _tall(tester);
      await harness.pump(
        tester,
        reply: _reply(
          access: accessJson(
            access: 'trial',
            daysLeft: 3,
            incomesRemaining: 2,
            expensesRemaining: 5,
          ),
        ),
      );

      expect(
        find.text('Prueba gratuita: quedan 3 días · 2 ingresos y 5 gastos'),
        findsOneWidget,
      );
      expect(find.text('Ver planes'), findsNothing);
    });

    testWidgets('trial, last day: singular', (tester) async {
      _tall(tester);
      await harness.pump(
        tester,
        reply: _reply(
          access: accessJson(
            access: 'trial',
            daysLeft: 1,
            incomesRemaining: 1,
            expensesRemaining: 1,
          ),
        ),
      );

      expect(
        find.text('Prueba gratuita: queda 1 día · 1 ingreso y 1 gasto'),
        findsOneWidget,
      );
    });

    testWidgets('read-only: figures stay; "Ver planes" with subscriptions', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(
        tester,
        modules: {...homeModules, AppModule.subscriptions},
        reply: _reply(access: accessJson(access: 'read_only')),
      );

      expect(
        find.text(
          'Modo lectura: los datos siguen visibles. Para registrar, '
          'contrate un plan.',
        ),
        findsOneWidget,
      );
      expect(find.byKey(const ValueKey('dashboard-hero')), findsOneWidget);

      await tester.tap(find.text('Ver planes'));
      await tester.pumpAndSettle();
      expect(find.text('Planes'), findsOneWidget);
    });

    testWidgets('read-only: a quick action opens the paywall', (tester) async {
      _tall(tester);
      await harness.pump(
        tester,
        reply: _reply(access: accessJson(access: 'read_only')),
      );

      await tester.tap(find.text('Registrar ingreso'));
      await tester.pumpAndSettle();

      expect(find.text('Tu cuenta es de solo consulta'), findsOneWidget);
      expect(harness.location, '/');
    });

    testWidgets('pro: no banner', (tester) async {
      _tall(tester);
      await harness.pump(tester);

      expect(find.byKey(const ValueKey('access-banner')), findsNothing);
    });
  });

  group('empty state and quick actions', () {
    testWidgets('no records: the empty state replaces hero and chart', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(
        tester,
        reply: _reply(
          isEmpty: true,
          monthIncome: 0,
          monthExpense: 0,
          access: accessJson(
            access: 'trial',
            daysLeft: 14,
            incomesRemaining: 20,
            expensesRemaining: 20,
          ),
        ),
      );

      expect(find.text(_emptyCopy), findsOneWidget);
      expect(find.byKey(const ValueKey('dashboard-hero')), findsNothing);
      expect(find.byKey(const ValueKey('trend-chart')), findsNothing);
      expect(find.byKey(const ValueKey('access-banner')), findsOneWidget);
      expect(
        find.widgetWithText(FilledButton, 'Registrar ingreso'),
        findsOneWidget,
      );
      expect(
        find.widgetWithText(OutlinedButton, 'Registrar gasto'),
        findsOneWidget,
      );
    });

    testWidgets('a quick action opens the create page; back, Home reloads', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(tester);

      await tester.tap(find.text('Registrar gasto'));
      await tester.pumpAndSettle();
      expect(harness.location, '/gastos/nuevo');

      await harness.back(tester);
      expect(harness.repository.calls, hasLength(2));
    });

    testWidgets('"Registrar ingreso" opens the income form', (tester) async {
      _tall(tester);
      await harness.pump(tester);

      await tester.tap(find.text('Registrar ingreso'));
      await tester.pumpAndSettle();

      expect(harness.location, '/ingresos/nuevo');
    });

    testWidgets('no active taxpayer: "Completar registro"', (tester) async {
      _tall(tester);
      await harness.pump(
        tester,
        held: [taxpayer(archivedAt: DateTime.utc(2026, 3))],
      );

      expect(harness.repository.calls, isEmpty);
      await tester.tap(find.text('Completar registro'));
      await tester.pumpAndSettle();

      expect(harness.location, '/contribuyentes/nuevo');
    });
  });

  group('load states', () {
    testWidgets('a skeleton while loading, never stale figures', (
      tester,
    ) async {
      final pending = Completer<Dashboard>();
      harness.repository.replies.add(pending);
      _tall(tester);
      await harness.pump(tester, settle: false);

      expect(find.byKey(const ValueKey('dashboard-skeleton')), findsOneWidget);
      expect(find.text('Total a pagar'), findsNothing);

      pending.complete(_reply());
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('dashboard-skeleton')), findsNothing);
    });

    testWidgets('a failure shows why and retries', (tester) async {
      harness.repository.replies.add(const ApiUnreachable());
      _tall(tester);
      await harness.pump(tester);

      expect(
        find.text('No hay conexión. Revisa tu internet e inténtalo de nuevo.'),
        findsOneWidget,
      );
      await tester.tap(find.text('Reintentar'));
      await tester.pumpAndSettle();

      expect(harness.repository.calls, hasLength(2));
      expect(find.byKey(const ValueKey('dashboard-hero')), findsOneWidget);
    });

    testWidgets('any other failure reads generic', (tester) async {
      harness.repository.replies.add(const ApiNotFound());
      _tall(tester);
      await harness.pump(tester);

      expect(find.text('No se pudo cargar el inicio.'), findsOneWidget);
    });

    testWidgets('pull to refresh reads again', (tester) async {
      // The default 800 × 600 surface: a pull is measured against it.
      await harness.pump(tester);

      await tester.fling(
        find.text('Total a pagar'),
        const Offset(0, 400),
        1000,
      );
      await tester.pumpAndSettle();

      expect(harness.repository.calls, hasLength(2));
    });

    testWidgets('switching taxpayer reloads for the new one', (tester) async {
      const other = '0192f0c4-0000-7000-8000-000000000002';
      _tall(tester);
      await harness.pump(
        tester,
        held: [
          taxpayer(regimes: [heldRegime('612')]),
          taxpayer(
            id: other,
            displayName: 'Beto',
            regimes: [heldRegime('612')],
            createdAt: DateTime.utc(2026, 2),
          ),
        ],
      );

      await harness.container
          .read(activeTaxpayerProvider.notifier)
          .select(other);
      await tester.pumpAndSettle();

      expect(harness.repository.calls.last, other);
    });
  });

  testWidgets('survives 200 % text on a small phone', (tester) async {
    tester.view
      ..physicalSize = const Size(360, 740)
      ..devicePixelRatio = 1;
    addTearDown(tester.view.reset);
    tester.platformDispatcher.textScaleFactorTestValue = 2;
    addTearDown(tester.platformDispatcher.clearTextScaleFactorTestValue);

    await harness.pump(
      tester,
      reply: _reply(
        access: accessJson(
          access: 'trial',
          daysLeft: 3,
          incomesRemaining: 2,
          expensesRemaining: 5,
        ),
      ),
    );
    await tester.scrollUntilVisible(
      find.text('Registrar gasto'),
      300,
      scrollable: find.byType(Scrollable).first,
    );
    await tester.pumpAndSettle();

    expect(tester.takeException(), isNull);
  });

  for (final brightness in Brightness.values) {
    testWidgets('meets contrast and target guidelines (${brightness.name})', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(
        tester,
        brightness: brightness,
        reply: _reply(
          access: accessJson(
            access: 'trial',
            daysLeft: 3,
            incomesRemaining: 2,
            expensesRemaining: 5,
          ),
        ),
      );

      await expectLater(tester, meetsGuideline(textContrastGuideline));
      await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
      await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
    });
  }
}
