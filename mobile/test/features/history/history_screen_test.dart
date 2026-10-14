import 'dart:async';

import 'package:control_contable/api/generated/models/history_year.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/history/presentation/history_screen.dart';
import 'package:control_contable/features/taxpayers/application/active_taxpayer.dart';
import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';
import 'package:flutter_test/flutter_test.dart';

import '../tax_engine/estimate_fixtures.dart';
import '../taxpayers/taxpayers_fakes.dart';
import 'history_fixtures.dart';
import 'pump_history.dart';

const String _recordsMarker =
    'Cambió desde que lo viste: agregaste o cambiaste registros o datos '
    'fiscales.';
const String _engineMarker =
    'Cambió desde que lo viste: se corrigió el cálculo.';

/// A tall phone, so the whole list is built without scrolling.
void _tall(WidgetTester tester) {
  tester.view
    ..physicalSize = const Size(400, 3000)
    ..devicePixelRatio = 1;
  addTearDown(tester.view.reset);
}

Finder _row(String period) => find.byKey(ValueKey('history-month-$period'));

Finder _inRow(String period, Finder finder) =>
    find.descendant(of: _row(period), matching: finder);

double _top(WidgetTester tester, Finder finder) =>
    tester.getTopLeft(finder.first).dy;

HistoryYear _year(
  List<Map<String, Object?>> months, {
  List<Map<String, Object?>> disclaimers = const [notTaxAdvice],
}) => historyFrom(historyYearJson(months: months, disclaimers: disclaimers));

void main() {
  late HistoryHarness harness;

  setUp(() => harness = HistoryHarness());

  group('module_history', () {
    testWidgets('OFF: no entry on Home and the location is not found', (
      tester,
    ) async {
      await harness.pump(
        tester,
        modules: historyModules.difference({AppModule.history}),
      );

      expect(find.byType(HistoryScreen), findsNothing);
      expect(harness.repository.calls, isEmpty);
      await harness.back(tester);
      expect(find.byTooltip('Historial'), findsNothing);
    });

    testWidgets('ON: Home offers "Historial", which opens the list', (
      tester,
    ) async {
      await harness.pump(tester, location: null);

      await tester.tap(find.byTooltip('Historial'));
      await tester.pumpAndSettle();

      expect(harness.location, '/historial');
      expect(find.byType(HistoryScreen), findsOneWidget);
    });
  });

  group('a fiscal year', () {
    testWidgets('asks the current year and lists months newest first', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(tester);

      expect(harness.repository.calls, [
        (taxpayerId: fixtureTaxpayerId, fiscalYear: 2026),
      ]);
      expect(find.text('Historial'), findsOneWidget);
      expect(
        _top(tester, _row('2026-03')),
        lessThan(_top(tester, _row('2026-02'))),
      );
      expect(
        _top(tester, _row('2026-02')),
        lessThan(_top(tester, _row('2026-01'))),
      );
    });

    testWidgets('each month shows its total, ISR, IVA and status chip', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(tester);

      expect(_inRow('2026-02', find.text('Febrero')), findsOneWidget);
      expect(_inRow('2026-02', find.text(r'$3,500.00')), findsOneWidget);
      expect(
        _inRow('2026-02', find.text(r'ISR $400.00 · IVA $3,100.00')),
        findsOneWidget,
      );
      expect(_inRow('2026-02', find.text('Completo')), findsOneWidget);
      expect(
        _inRow('2026-02', find.text('Vence el 17 de marzo de 2026')),
        findsOneWidget,
      );
    });

    testWidgets('the current month is tagged and never marked', (tester) async {
      _tall(tester);
      await harness.pump(
        tester,
        reply: _year([
          historyMonthJson(
            '2026-03',
            isCurrent: true,
            viewedAt: '2026-03-10T17:00:00Z',
          ),
        ]),
      );

      expect(_inRow('2026-03', find.text('Mes en curso')), findsOneWidget);
      expect(_inRow('2026-03', find.textContaining('Cambió')), findsNothing);
      expect(_inRow('2026-03', find.textContaining('Visto')), findsNothing);
    });

    testWidgets('a late record marks the month in plain words', (tester) async {
      _tall(tester);
      await harness.pump(tester);

      expect(_inRow('2026-02', find.text(_recordsMarker)), findsOneWidget);
      expect(
        _inRow('2026-02', find.byIcon(Icons.update_rounded)),
        findsOneWidget,
      );
    });

    testWidgets('an engine correction says the calculation was corrected', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(
        tester,
        reply: _year([
          historyMonthJson(
            '2026-02',
            viewedAt: '2026-03-03T17:00:00Z',
            changeReason: 'engine',
          ),
        ]),
      );

      expect(_inRow('2026-02', find.text(_engineMarker)), findsOneWidget);
    });

    testWidgets('a viewed, unchanged month says when it was seen', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(tester);

      expect(_inRow('2026-01', find.text('Visto el 14 feb')), findsOneWidget);
      expect(_inRow('2026-01', find.textContaining('Cambió')), findsNothing);
    });

    testWidgets('a never-viewed month carries no marker', (tester) async {
      _tall(tester);
      await harness.pump(tester, reply: _year([historyMonthJson('2026-02')]));

      expect(_inRow('2026-02', find.textContaining('Cambió')), findsNothing);
      expect(_inRow('2026-02', find.textContaining('Visto')), findsNothing);
    });

    testWidgets('a month with no regime says so, with no amount', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(
        tester,
        reply: _year([
          historyMonthJson('2026-02'),
          historyMonthJson(
            '2026-01',
            status: 'none',
            hasRecords: false,
            tax: null,
          ),
        ]),
      );

      expect(
        _inRow('2026-01', find.text('Sin régimen en este mes')),
        findsOneWidget,
      );
      expect(_inRow('2026-01', find.textContaining(r'$')), findsNothing);
    });

    testWidgets('a month with no records keeps the engine figure', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(
        tester,
        reply: _year([
          historyMonthJson('2026-02', hasRecords: false),
          historyMonthJson('2026-01'),
        ]),
      );

      expect(
        _inRow('2026-02', find.text('Sin registros en este mes')),
        findsOneWidget,
      );
      expect(_inRow('2026-02', find.text(r'$2,284.29')), findsOneWidget);
    });

    testWidgets('IVA in favor reads as saldo a favor; withholdings apart', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(
        tester,
        reply: _year([
          historyMonthJson(
            '2026-02',
            tax: historyTaxJson(
              iva: 0,
              ivaOutcome: 'in_favor',
              ivaInFavor: 92400,
              withholdings: 1033333,
            ),
          ),
        ]),
      );

      expect(
        _inRow(
          '2026-02',
          find.text(r'ISR $150.96 · IVA: saldo a favor $924.00'),
        ),
        findsOneWidget,
      );
      expect(
        _inRow('2026-02', find.text(r'Retenciones a enterar $10,333.33')),
        findsOneWidget,
      );
    });

    testWidgets('year totals, then the server legend, then the months', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(tester);

      final totals = find.byKey(const ValueKey('history-year-totals'));
      expect(
        find.descendant(
          of: totals,
          matching: find.text('Suma de estimados mensuales'),
        ),
        findsOneWidget,
      );
      expect(
        find.descendant(of: totals, matching: find.text(r'$6,864.29')),
        findsOneWidget,
      );
      expect(
        find.descendant(
          of: totals,
          matching: find.text('No es la declaración anual.'),
        ),
        findsOneWidget,
      );
      final legend = find.text(legendText);
      expect(legend, findsOneWidget);
      expect(_top(tester, legend), greaterThan(_top(tester, totals)));
      expect(_top(tester, legend), lessThan(_top(tester, _row('2026-03'))));
      expect(find.text(notTaxAdvice['text']! as String), findsOneWidget);
    });

    testWidgets('a month pending CPA review raises the strong banner on top', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(
        tester,
        reply: _year(
          [historyMonthJson('2026-02', cpaReviewPending: true)],
          disclaimers: [notTaxAdvice, cpaReviewDisclaimer],
        ),
      );

      final banner = find.byKey(const ValueKey('history-cpa-banner'));
      expect(banner, findsOneWidget);
      expect(
        _top(tester, banner),
        lessThan(
          _top(tester, find.byKey(const ValueKey('history-year-totals'))),
        ),
      );
    });
  });

  group('choosing the year', () {
    testWidgets('2026 offers only 2026', (tester) async {
      await harness.pump(tester);

      await tester.tap(find.byKey(const ValueKey('history-year')));
      await tester.pumpAndSettle();

      expect(find.text('2026'), findsWidgets);
      expect(find.text('2025'), findsNothing);
    });

    testWidgets('a past year is read again from the server', (tester) async {
      harness.now = DateTime.utc(2027, 2, 10, 18);
      await harness.pump(tester);

      await tester.tap(find.byKey(const ValueKey('history-year')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('2026').last);
      await tester.pumpAndSettle();

      expect(harness.repository.calls.map((c) => c.fiscalYear), [2027, 2026]);
    });
  });

  group('opening a month', () {
    testWidgets('opens the estimate of that month and marks it viewed', (
      tester,
    ) async {
      _tall(tester);
      harness.estimates.fallback = estimateFrom(
        estimateJson(period: '2026-02', dueDate: '2026-03-17'),
      );
      await harness.pump(tester);

      await tester.tap(_row('2026-02'));
      await tester.pumpAndSettle();

      expect(harness.location, '/estimado?periodo=2026-02');
      expect(harness.estimates.calls.last, (
        taxpayerId: fixtureTaxpayerId,
        period: '2026-02',
      ));
      final viewed = harness.repository.viewed.single;
      expect(viewed.taxpayerId, fixtureTaxpayerId);
      expect(viewed.period, '2026-02');
      expect(viewed.request.figuresFingerprint, fixtureFiguresFingerprint);
    });

    testWidgets('returning reads the year again', (tester) async {
      _tall(tester);
      await harness.pump(tester);

      await tester.tap(_row('2026-01'));
      await tester.pumpAndSettle();
      harness.repository.fallback = _year([
        historyMonthJson('2026-01', viewedAt: '2026-03-20T18:00:00Z'),
      ]);
      await harness.back(tester);

      expect(harness.repository.calls.length, greaterThanOrEqualTo(2));
      expect(_inRow('2026-01', find.text('Visto el 20 mar')), findsOneWidget);
    });

    testWidgets('returning announces the month\'s new state politely (§6)', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(tester);
      await tester.tap(_row('2026-02'));
      await tester.pumpAndSettle();
      harness.repository.fallback = _year([
        historyMonthJson('2026-02', viewedAt: '2026-03-20T18:00:00Z'),
      ]);
      tester.takeAnnouncements();

      await harness.back(tester);

      final announced = tester.takeAnnouncements();
      expect(announced.map((a) => a.message), [
        'Febrero de 2026. Visto el 20 mar.',
      ]);
      expect(announced.single.assertiveness, Assertiveness.polite);
    });

    testWidgets('a failed re-read on return announces nothing', (tester) async {
      _tall(tester);
      await harness.pump(tester);
      await tester.tap(_row('2026-02'));
      await tester.pumpAndSettle();
      harness.repository.replies.add(const ApiUnexpected());
      tester.takeAnnouncements();

      await harness.back(tester);

      expect(
        tester.takeAnnouncements().map((a) => a.message),
        isNot(contains(startsWith('Febrero de 2026'))),
      );
    });

    testWidgets('the row is one target read in order, marker included', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(tester);

      expect(
        find.bySemanticsLabel(
          'Febrero de 2026. Total a pagar: 3,500 pesos. Completo. '
          '$_recordsMarker',
        ),
        findsOneWidget,
      );
      expect(
        find.bySemanticsLabel(
          RegExp(r'^Marzo de 2026, mes en curso\. Total a pagar: 2,284 pesos'),
        ),
        findsOneWidget,
      );
    });
  });

  group('links to records (US-H6)', () {
    testWidgets('"Ver ingresos" opens that month\'s incomes', (tester) async {
      _tall(tester);
      await harness.pump(tester);

      await tester.tap(_inRow('2026-01', find.text('Ver ingresos')));
      await tester.pumpAndSettle();

      expect(harness.location, '/ingresos?periodo=2026-01');
      expect(harness.incomes.calls.last, 'list:2026-01 * -');
      expect(harness.repository.viewed, isEmpty);
    });

    testWidgets('"Ver gastos" opens that month\'s expenses', (tester) async {
      _tall(tester);
      await harness.pump(tester);

      await tester.tap(_inRow('2026-02', find.text('Ver gastos')));
      await tester.pumpAndSettle();

      expect(harness.location, '/gastos?periodo=2026-02');
      expect(harness.expenses.calls.last, 'list:2026-02 * -');
    });

    testWidgets('each link is its own target, named with its month', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(tester);

      expect(
        find.bySemanticsLabel('Ver ingresos de febrero de 2026'),
        findsOneWidget,
      );
      expect(
        find.bySemanticsLabel('Ver gastos de febrero de 2026'),
        findsOneWidget,
      );
    });

    testWidgets('a link hides while its module is off', (tester) async {
      _tall(tester);
      await harness.pump(
        tester,
        modules: historyModules.difference({AppModule.expenses}),
      );

      expect(_inRow('2026-01', find.text('Ver ingresos')), findsOneWidget);
      expect(find.text('Ver gastos'), findsNothing);
    });

    testWidgets('with both record modules off, no link at all', (tester) async {
      _tall(tester);
      await harness.pump(
        tester,
        modules: historyModules.difference({
          AppModule.incomes,
          AppModule.expenses,
        }),
      );

      expect(find.text('Ver ingresos'), findsNothing);
      expect(find.text('Ver gastos'), findsNothing);
    });
  });

  group('states', () {
    testWidgets('a skeleton shows while the year loads', (tester) async {
      final pending = Completer<HistoryYear>();
      harness.repository.replies.add(pending);
      await harness.pump(tester, settle: false);

      expect(find.byKey(const ValueKey('history-skeleton')), findsOneWidget);
      expect(find.text('Cargando el historial…'), findsOneWidget);

      pending.complete(historyFrom(historyYearJson()));
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('history-skeleton')), findsNothing);
    });

    testWidgets('a failure offers a retry', (tester) async {
      harness.repository.replies.add(const ApiUnexpected());
      await harness.pump(tester);

      expect(find.text('No se pudo cargar el historial.'), findsOneWidget);
      await tester.tap(find.text('Reintentar'));
      await tester.pumpAndSettle();

      expect(harness.repository.calls, hasLength(2));
      expect(_row('2026-03'), findsOneWidget);
    });

    testWidgets('pulling down reads the year again', (tester) async {
      await harness.pump(tester);

      await tester.fling(
        find.text('Ejercicio fiscal'),
        const Offset(0, 400),
        1000,
      );
      await tester.pumpAndSettle();

      expect(harness.repository.calls, hasLength(2));
    });

    testWidgets('an unreachable server says so', (tester) async {
      harness.repository.replies.add(const ApiUnreachable());
      await harness.pump(tester);

      expect(
        find.text('No hay conexión. Revisa tu internet e inténtalo de nuevo.'),
        findsOneWidget,
      );
    });

    testWidgets('warnings sit with the figures, apart from the notes', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(
        tester,
        reply: _year(
          threeMonthHistory(),
          disclaimers: [notTaxAdvice, provisionalParametersDisclaimer],
        ),
      );

      final warning = find.text(
        provisionalParametersDisclaimer['text']! as String,
      );
      expect(warning, findsOneWidget);
      expect(_top(tester, warning), lessThan(_top(tester, _row('2026-03'))));
    });

    testWidgets('no records at all: the empty message and a way to start', (
      tester,
    ) async {
      await harness.pump(
        tester,
        reply: _year([
          historyMonthJson('2026-03', isCurrent: true, hasRecords: false),
          historyMonthJson('2026-02', hasRecords: false),
        ]),
      );

      expect(
        find.text(
          'Aún no hay registros. El historial se llena con su primer '
          'ingreso o gasto.',
        ),
        findsOneWidget,
      );
      expect(_row('2026-02'), findsNothing);
      await tester.tap(find.text('Registrar ingreso'));
      await tester.pumpAndSettle();
      expect(harness.location, '/ingresos/nuevo');
    });

    testWidgets('switching taxpayer never shows the previous figures', (
      tester,
    ) async {
      const other = '0192f0c4-0000-7000-8000-000000000002';
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
      final shown = harness.repository.calls.single.taxpayerId;
      final next = shown == other ? fixtureTaxpayerId : other;
      final pending = Completer<HistoryYear>();
      harness.repository.replies.add(pending);

      unawaited(
        harness.container.read(activeTaxpayerProvider.notifier).select(next),
      );
      for (var frame = 0; frame < 5; frame++) {
        await tester.pump();
      }

      expect(find.byKey(const ValueKey('history-skeleton')), findsOneWidget);
      expect(_row('2026-02'), findsNothing);
      pending.complete(historyFrom(historyYearJson(taxpayerId: next)));
      await tester.pumpAndSettle();
      expect(harness.repository.calls.last.taxpayerId, next);
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
      reply: _year([
        historyMonthJson(
          '2026-02',
          status: 'incomplete',
          tax: historyTaxJson(
            isr: 3600000,
            iva: 12000000,
            withholdings: 1033333,
          ),
          viewedAt: '2026-03-03T17:00:00Z',
          changeReason: 'records',
        ),
      ]),
    );
    await tester.scrollUntilVisible(
      find.text(_recordsMarker),
      200,
      scrollable: find
          .descendant(
            of: find.byType(HistoryScreen),
            matching: find.byType(Scrollable),
          )
          .first,
    );

    expect(tester.takeException(), isNull);
  });

  for (final brightness in Brightness.values) {
    testWidgets('meets contrast and target guidelines (${brightness.name})', (
      tester,
    ) async {
      _tall(tester);
      await harness.pump(tester, brightness: brightness);

      await expectLater(tester, meetsGuideline(textContrastGuideline));
      await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
      await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
    });
  }
}
