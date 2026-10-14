import 'dart:ui' show Tristate;

import 'package:control_contable/core/flags/app_module.dart';
import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_test/flutter_test.dart';

import 'dashboard_fixtures.dart';
import 'pump_dashboard.dart';

/// What TalkBack / VoiceOver read for the tax chart of the default reply.
const String _taxSummary =
    'Impuestos estimados, últimos 3 meses. Marzo: 2,284 pesos con 29 '
    'centavos. Mes más alto: febrero, 3,500 pesos.';

/// What they read for the income/expense chart of the default reply.
const String _incomeExpenseSummary =
    'Ingresos y gastos, últimos 3 meses. Marzo: ingresos cobrados 30,000 '
    'pesos, gastos pagados 10,000 pesos. En el periodo: ingresos cobrados '
    '65,000 pesos, gastos pagados 20,000 pesos.';

/// A tall phone, so the whole Home is built without scrolling.
void _tall(WidgetTester tester) {
  tester.view
    ..physicalSize = const Size(400, 3000)
    ..devicePixelRatio = 1;
  addTearDown(tester.view.reset);
}

/// The node a screen reader focuses for the chart: the one its summary
/// labels.
SemanticsNode _summaryNode(WidgetTester tester) =>
    tester.getSemantics(find.byKey(const ValueKey('trend-chart')));

/// The node a screen reader focuses for one month's bar.
SemanticsNode _barNode(WidgetTester tester, String period) =>
    tester.getSemantics(find.byKey(ValueKey('trend-bar-$period')));

List<SemanticsNode> _children(SemanticsNode node) {
  final out = <SemanticsNode>[];
  node.visitChildren((child) {
    out.add(child);
    return true;
  });
  return out;
}

/// The nodes under [node] that a screen reader can land on (a label), in
/// tree order, through unlabeled wrappers such as the scroll view's.
List<SemanticsNode> _labeledDescendants(SemanticsNode node) => [
  for (final child in _children(node))
    if (child.label.isNotEmpty) child else ..._labeledDescendants(child),
];

/// M6-07 (tech-lead gate, part C): the QA could not drive TalkBack through
/// adb (M6-06 §6), so the chart's accessibility is pinned on the semantics
/// tree the platform screen readers read: the summary is one node of its
/// own whose label is exactly the spoken sentence (MODULE-dashboard §6),
/// each bar is a separate, actionable button that says its month and
/// values, and the table alternative reads every month as its own block.
void main() {
  late DashboardHarness harness;

  setUp(() => harness = DashboardHarness());

  testWidgets('the chart summary is its own node and says exactly the '
      'spoken summary, nothing merged into it', (tester) async {
    final handle = tester.ensureSemantics();
    _tall(tester);
    await harness.pump(tester);

    final summary = _summaryNode(tester);
    expect(summary.label, _taxSummary);
    final data = summary.getSemanticsData();
    expect(data.flagsCollection.isButton, isFalse, reason: 'not a bar');
    expect(data.hasAction(SemanticsAction.tap), isFalse);
    expect(_labeledDescendants(summary).map((n) => n.label), [
      startsWith('enero de 2026: '),
      startsWith('febrero de 2026: '),
      startsWith('marzo de 2026: '),
    ], reason: 'one child node per bar, oldest first (reading order)');
    handle.dispose();
  });

  testWidgets('each bar is a selectable button that announces its month, '
      'values, total and status', (tester) async {
    final handle = tester.ensureSemantics();
    _tall(tester);
    await harness.pump(tester);

    final february = _barNode(tester, '2026-02').getSemanticsData();
    expect(
      february.label,
      'febrero de 2026: IVA 3,100 pesos, ISR 400 pesos, total 3,500 pesos, '
      'incompleto',
    );
    expect(february.flagsCollection.isButton, isTrue);
    expect(february.flagsCollection.isSelected, Tristate.isFalse);
    expect(
      _barNode(tester, '2026-03').getSemanticsData().flagsCollection.isSelected,
      Tristate.isTrue,
      reason: 'the current month starts selected',
    );
    handle.dispose();
  });

  // QA M6-06 D-1: excludeSemantics dropped the GestureDetector's tap, so
  // TalkBack announced "button" over a node it could not activate (the
  // same defect as the taxpayer switcher in QA M3/M4 D1).
  testWidgets('a screen reader selects a month through the tap action of '
      'its bar', (tester) async {
    final handle = tester.ensureSemantics();
    _tall(tester);
    await harness.pump(tester);

    final bar = _barNode(tester, '2026-02');
    expect(bar.getSemanticsData().hasAction(SemanticsAction.tap), isTrue);
    tester.semantics.tap(find.semantics.byLabel(RegExp('^febrero de 2026: ')));
    await tester.pumpAndSettle();

    expect(
      _barNode(tester, '2026-02').getSemanticsData().flagsCollection.isSelected,
      Tristate.isTrue,
    );
    expect(
      find.descendant(
        of: find.byKey(const ValueKey('trend-selection')),
        matching: find.text('febrero de 2026'),
      ),
      findsOneWidget,
    );
    handle.dispose();
  });

  // A window of one month (period 2026-01): a lone bar must not merge into
  // the summary, or the summary would read as a button and lose its words.
  testWidgets('a one-month window keeps the summary and the bar apart', (
    tester,
  ) async {
    final handle = tester.ensureSemantics();
    _tall(tester);
    final trend = [
      trendPointJson(
        '2026-01',
        income: 1000000,
        expense: 400000,
        tax: trendTaxJson(iva: 96000, isr: 12000),
      ),
    ];
    await harness.pump(
      tester,
      reply: dashboardFrom(dashboardJson(trend: trend)),
    );

    final summary = _summaryNode(tester);
    final bar = _barNode(tester, '2026-01');
    expect(bar.id, isNot(summary.id));
    expect(summary.label, startsWith('Impuestos estimados, '));
    expect(summary.label, isNot(contains('enero de 2026:')));
    expect(summary.getSemanticsData().flagsCollection.isButton, isFalse);
    expect(bar.label, startsWith('enero de 2026: IVA 960 pesos'));
    handle.dispose();
  });

  testWidgets('the income/expense view says both series for the month and '
      'the window totals', (tester) async {
    final handle = tester.ensureSemantics();
    _tall(tester);
    await harness.pump(tester);

    await tester.tap(find.text('Ingresos y gastos'));
    await tester.pumpAndSettle();

    expect(_summaryNode(tester).label, _incomeExpenseSummary);
    expect(
      _barNode(tester, '2026-03').label,
      'marzo de 2026: ingresos cobrados 30,000 pesos, gastos pagados 10,000 '
      'pesos',
    );
    handle.dispose();
  });

  testWidgets('engine OFF: the summary speaks income and expense only, '
      'never a tax', (tester) async {
    final handle = tester.ensureSemantics();
    _tall(tester);
    await harness.pump(
      tester,
      modules: homeModules.difference({AppModule.taxEngine}),
      reply: dashboardFrom(dashboardJson(engine: false)),
    );

    final summary = _summaryNode(tester).label;
    expect(summary, _incomeExpenseSummary);
    expect(
      summary,
      isNot(matches(RegExp('impuesto|IVA|ISR', caseSensitive: false))),
    );
    handle.dispose();
  });

  testWidgets('the table reads each month as its own block: a header, '
      'then one sentence per figure', (tester) async {
    final handle = tester.ensureSemantics();
    _tall(tester);
    await harness.pump(tester);

    await tester.tap(find.text('Ver como tabla'));
    await tester.pumpAndSettle();

    final february = tester.getSemantics(
      find.descendant(
        of: find.byKey(const ValueKey('trend-table')),
        matching: find.text('febrero de 2026'),
      ),
    );
    // The month's block is one heading node ("febrero de 2026, Incompleto")
    // whose children are its figures, one sentence each.
    expect(february.label, 'febrero de 2026\nIncompleto');
    expect(february.getSemanticsData().flagsCollection.isHeader, isTrue);
    expect(
      _children(february).map((n) => n.label),
      containsAllInOrder([
        'Ingresos cobrados: 25,000 pesos',
        'Gastos pagados: 6,000 pesos',
        'IVA: 3,100 pesos',
        'ISR: 400 pesos',
        'Retenciones: 0 pesos',
        'Total: 3,500 pesos',
      ]),
    );
    expect(
      find.descendant(
        of: find.byKey(const ValueKey('trend-table')),
        matching: find.text('Incompleto'),
      ),
      findsOneWidget,
      reason: 'the status is text, not colour',
    );
    handle.dispose();
  });
}
