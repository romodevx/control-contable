import 'package:control_contable/api/generated/models/annual_computed_section_kind.dart';
import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/annual/annual_routes.dart';
import 'package:control_contable/features/annual/presentation/annual_breakdown_screen.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../tax_engine/estimate_fixtures.dart' show fixtureTaxpayerId;
import '../taxpayers/taxpayers_fakes.dart' show detail;
import 'annual_fixtures.dart';
import 'pump_annual.dart';

const Size _phone = Size(360, 740);

final String _pfBreakdown = AnnualRoutes.breakdown(
  fixtureTaxpayerId,
  2026,
  AnnualComputedSectionKind.pfArt152,
);

Finder _line(int seq) => find.byKey(ValueKey('line-$seq'));

Finder _inLine(int seq, Finder finder) =>
    find.descendant(of: _line(seq), matching: finder);

Future<void> _reveal(WidgetTester tester, Finder finder) async {
  await tester.scrollUntilVisible(
    finder,
    200,
    scrollable: find
        .descendant(
          of: find.byType(AnnualBreakdownScreen),
          matching: find.byType(Scrollable),
        )
        .first,
  );
  await tester.pumpAndSettle();
}

Future<void> _open(
  AnnualHarness harness,
  WidgetTester tester, {
  Object? reply,
}) async {
  if (reply != null) harness.repository.breakdownFallback = reply;
  await harness.pump(tester, location: _pfBreakdown);
}

void main() {
  late AnnualHarness harness;

  setUp(() => harness = AnnualHarness());

  testWidgets('asks for the section and year; shows the header and legend', (
    tester,
  ) async {
    await _open(harness, tester);

    expect(harness.repository.breakdownCalls, [
      (
        taxpayerId: fixtureTaxpayerId,
        fiscalYear: 2026,
        section: AnnualComputedSectionKind.pfArt152,
      ),
    ]);
    expect(find.text('Cómo se calculó'), findsOneWidget);
    expect(find.text('Personas físicas: declaración anual'), findsOneWidget);
    expect(find.text('Ejercicio 2026 · Regímenes 612'), findsOneWidget);
    expect(find.text('Completo'), findsOneWidget);
    expect(find.text('Parámetros 2026 v1 · Motor 1.1.0'), findsOneWidget);
    expect(find.text(annualLegendText), findsOneWidget);
  });

  testWidgets('lines show in seq order, whatever order they arrive in', (
    tester,
  ) async {
    await _open(
      harness,
      tester,
      reply: annualBreakdownFrom(
        annualBreakdownJson(lines: an01Lines.reversed.toList()),
      ),
    );

    final tops = [
      for (final seq in [1, 2, 3]) tester.getTopLeft(_line(seq)).dy,
    ];
    expect(tops, orderedEquals([...tops]..sort()));
  });

  testWidgets('amounts are es-MX centavos; a subtraction shows and reads as '
      'one; the rate keeps its decimals', (tester) async {
    await _open(harness, tester);

    expect(_inLine(1, find.text(r'$500,000.00')), findsOneWidget);
    expect(_inLine(3, find.text(r'−$40,000.00')), findsOneWidget);
    expect(
      find.bySemanticsLabel(
        RegExp(r'^Deducciones personales: menos 40,000 pesos$'),
      ),
      findsOneWidget,
    );
    expect(_inLine(4, find.text('Tasa: 30\u00a0%')), findsOneWidget);
  });

  testWidgets('each figure says where it comes from, payments by source', (
    tester,
  ) async {
    await _open(harness, tester);

    expect(_inLine(1, find.text('De tus registros')), findsOneWidget);
    expect(_inLine(1, find.text('Régimen 612')), findsOneWidget);
    expect(
      _inLine(3, find.text('De tus deducciones personales')),
      findsOneWidget,
    );
    expect(_inLine(4, find.text('Tarifa oficial')), findsOneWidget);
    await _reveal(tester, _line(7));
    expect(_inLine(5, find.text('Pago presentado')), findsOneWidget);
    expect(_inLine(6, find.text('Estimado por la app')), findsOneWidget);
    expect(
      _inLine(7, find.text('De tus constancias de sueldos')),
      findsOneWidget,
    );
  });

  testWidgets('a cap that binds is its own "Tope aplicado" line', (
    tester,
  ) async {
    await _open(harness, tester);

    final cap = _line(2);
    expect(
      find.descendant(
        of: cap,
        matching: find.text(
          r'Tope aplicado: colegiaturas, de $30,000.00 a $12,900.00',
        ),
      ),
      findsOneWidget,
    );
    expect(
      find.descendant(of: cap, matching: find.byIcon(Icons.info_outline)),
      findsOneWidget,
    );
  });

  testWidgets('legal basis opens; a CPA-unverified line is marked', (
    tester,
  ) async {
    await _open(harness, tester);

    await tester.tap(_inLine(1, find.text('Fundamento: LISR 152')));
    await tester.pumpAndSettle();
    expect(
      _inLine(1, find.text('Ley del Impuesto sobre la Renta, artículo 152')),
      findsOneWidget,
    );
    await _reveal(tester, _line(8));
    expect(
      _inLine(8, find.text('Sin validar por un contador')),
      findsOneWidget,
    );
  });

  testWidgets('strong banner, warnings and notes travel with the figures', (
    tester,
  ) async {
    await _open(
      harness,
      tester,
      reply: annualBreakdownFrom(
        annualBreakdownJson(
          section: resicoSectionJson(),
          disclaimers: [
            annualEstimateDisclaimer,
            annualCpaDisclaimer,
            paymentsEstimatedDisclaimer,
          ],
        ),
      ),
    );

    expect(find.byKey(const ValueKey('cpa-review-banner')), findsOneWidget);
    expect(
      find.text(
        'Pendiente de validar por un contador: la renovación de la facilidad '
        'de RESICO en la RMF 2027.',
      ),
      findsOneWidget,
    );
    expect(
      find.byKey(const ValueKey('disclaimer-payments_estimated')),
      findsOneWidget,
    );
    await _reveal(
      tester,
      find.byKey(const ValueKey('disclaimer-annual_estimate')),
    );
  });

  testWidgets('an incomplete section names what is missing', (tester) async {
    await _open(
      harness,
      tester,
      reply: annualBreakdownFrom(
        annualBreakdownJson(
          section: sectionJson(
            status: 'incomplete',
            missingInputs: [
              missingJson(
                'salary_statement',
                'Constancia de sueldos',
                'salary_statements',
              ),
            ],
          ),
        ),
      ),
    );

    final named = find.text('Constancia de sueldos');
    await _reveal(tester, named);
    expect(find.text('Falta:'), findsOneWidget);
    // Its screen ships in ANN-20: named, never a link to a dead end.
    expect(
      find.ancestor(of: named, matching: find.byType(ListTile)),
      findsNothing,
    );
  });

  testWidgets('a section the year does not have says so, with a retry', (
    tester,
  ) async {
    harness.repository.breakdownReplies.add(
      ApiRejected(
        code: ErrorCode.businessRuleViolation,
        details: [detail('section', 'section_not_present')],
      ),
    );
    await _open(harness, tester);

    expect(
      find.text('Tu estimación de ese año no tiene esta sección.'),
      findsOneWidget,
    );
    await tester.tap(find.text('Reintentar'));
    await tester.pumpAndSettle();
    expect(find.text('Cómo se calculó'), findsOneWidget);
    expect(harness.repository.breakdownCalls, hasLength(2));
  });

  testWidgets('an unknown section in the location is not found', (
    tester,
  ) async {
    await harness.pump(
      tester,
      location: '/anual/$fixtureTaxpayerId/2026/secciones/otra',
    );

    expect(harness.repository.breakdownCalls, isEmpty);
    expect(find.text('Página no encontrada'), findsOneWidget);
  });

  testWidgets('survives 200 % text on a phone', (tester) async {
    tester.view
      ..physicalSize = _phone
      ..devicePixelRatio = 1;
    addTearDown(tester.view.reset);
    tester.platformDispatcher.textScaleFactorTestValue = 2;
    addTearDown(tester.platformDispatcher.clearTextScaleFactorTestValue);

    await _open(harness, tester);
    await _reveal(tester, _line(8));

    expect(tester.takeException(), isNull);
  });

  for (final brightness in Brightness.values) {
    testWidgets('meets contrast and target guidelines (${brightness.name})', (
      tester,
    ) async {
      await harness.pump(
        tester,
        brightness: brightness,
        location: _pfBreakdown,
      );

      await expectLater(tester, meetsGuideline(textContrastGuideline));
      await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
      await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
    });
  }
}
