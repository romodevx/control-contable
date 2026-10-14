import 'dart:async';
import 'dart:ui' show Tristate;

import 'package:control_contable/api/generated/models/breakdown.dart';
import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/api/generated/models/field_error.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/features/tax_engine/presentation/breakdown_screen.dart';
import 'package:control_contable/features/tax_engine/tax_engine_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'breakdown_fixtures.dart';
import 'estimate_fixtures.dart';
import 'pump_estimate.dart';

final String _breakdown612 = TaxEngineRoutes.breakdown(
  fixtureTaxpayerId,
  '2026-03',
  '612',
);

/// A small phone in portrait (logical pixels).
const Size _phone = Size(360, 740);

Finder _line(int seq) => find.byKey(ValueKey('line-$seq'));

Finder _inLine(int seq, Finder finder) =>
    find.descendant(of: _line(seq), matching: finder);

Finder _inGroup(String tax, Finder finder) =>
    find.descendant(of: find.byKey(ValueKey('group-$tax')), matching: finder);

Future<void> _reveal(WidgetTester tester, Finder finder) async {
  await tester.scrollUntilVisible(
    finder,
    200,
    scrollable: find
        .descendant(
          of: find.byType(BreakdownScreen),
          matching: find.byType(Scrollable),
        )
        .first,
  );
  await tester.pumpAndSettle();
}

/// Opens the 612 breakdown of March 2026 answering [reply].
Future<void> _open(
  EstimateHarness harness,
  WidgetTester tester, {
  Object? reply,
}) async {
  if (reply != null) harness.repository.breakdownFallback = reply;
  await harness.pump(tester, location: _breakdown612);
}

Breakdown _incomplete() => breakdownFrom(
  breakdownJson(
    regime: regimeJson(
      '612',
      status: 'incomplete',
      isrCents: 415096,
      missingInputs: [
        {
          'key': 'prior_provisional_payments',
          'label': 'Pagos provisionales anteriores del ejercicio',
          'months': ['2026-01', '2026-02'],
        },
      ],
    ),
    disclaimers: [notTaxAdvice, incompleteInputsDisclaimer],
  ),
);

void main() {
  late EstimateHarness harness;

  setUp(() => harness = EstimateHarness());

  testWidgets('asks the server for the regime and month and shows the '
      'header', (tester) async {
    await _open(harness, tester);

    expect(harness.repository.breakdownCalls, [
      (taxpayerId: fixtureTaxpayerId, period: '2026-03', regime: '612'),
    ]);
    expect(find.text('Desglose'), findsOneWidget);
    expect(
      find.text('Actividades Empresariales y Profesionales'),
      findsOneWidget,
    );
    expect(find.text('Régimen 612 · marzo de 2026'), findsOneWidget);
    expect(find.text('Completo'), findsOneWidget);
    expect(find.text('Parámetros 2026 v1 · Motor 1.0.0'), findsOneWidget);
  });

  testWidgets('lines show in seq order, whatever order they arrive in', (
    tester,
  ) async {
    await _open(
      harness,
      tester,
      reply: breakdownFrom(breakdownJson(lines: isrLines612.reversed.toList())),
    );

    final tops = [
      for (final seq in [1, 2, 3]) tester.getTopLeft(_line(seq)).dy,
    ];
    expect(tops, orderedEquals([...tops]..sort()));
  });

  testWidgets('amounts are es-MX centavos; a subtraction shows and reads '
      'as one', (tester) async {
    await _open(harness, tester);

    expect(_inLine(1, find.text(r'$90,000.00')), findsOneWidget);
    expect(_inLine(2, find.text(r'−$30,000.00')), findsOneWidget);
    expect(
      find.bySemanticsLabel(
        RegExp(r'^Deducciones pagadas acumuladas ene–mar: menos 30,000 pesos$'),
      ),
      findsOneWidget,
    );
    expect(
      find.bySemanticsLabel(
        RegExp(r'^ISR a pagar: 150 pesos con 96 centavos$'),
      ),
      findsOneWidget,
    );
  });

  testWidgets('a multiplication shows its rate without float rounding', (
    tester,
  ) async {
    await _open(harness, tester);

    expect(_inLine(5, find.text('Tasa: 21.36 %')), findsOneWidget);
  });

  testWidgets('each line says where its figure comes from', (tester) async {
    await _open(harness, tester);

    expect(_inLine(1, find.text('De tus registros')), findsOneWidget);
    expect(_inLine(3, find.text('Calculado')), findsOneWidget);
    expect(_inLine(4, find.text('Tarifa oficial')), findsOneWidget);
    await _reveal(tester, _line(7));
    expect(_inLine(7, find.text('Dato que capturaste')), findsOneWidget);
  });

  group('legal basis', () {
    testWidgets('each cited line names its articles, collapsed', (
      tester,
    ) async {
      await _open(harness, tester);

      expect(
        _inLine(2, find.text('Fundamento: LISR 103, LISR 105')),
        findsOneWidget,
      );
      expect(
        find.text('Ley del Impuesto sobre la Renta, artículo 103'),
        findsNothing,
      );
      final toggle = tester.getSemantics(
        _inLine(2, find.text('Fundamento: LISR 103, LISR 105')),
      );
      expect(toggle.flagsCollection.isExpanded, Tristate.isFalse);
    });

    testWidgets('expands to the article, its note and a link to the law', (
      tester,
    ) async {
      await _open(harness, tester);
      await _reveal(tester, _line(8));

      await tester.tap(_inLine(8, find.text('Fundamento: LISR 106')));
      await tester.pumpAndSettle();

      expect(
        _inLine(
          8,
          find.text(
            'Ley del Impuesto sobre la Renta, artículo 106, '
            'párrafo quinto',
          ),
        ),
        findsOneWidget,
      );
      expect(
        tester
            .getSemantics(_inLine(8, find.text('Fundamento: LISR 106')))
            .flagsCollection
            .isExpanded,
        Tristate.isTrue,
      );
      await tester.tap(
        _inLine(8, find.text('Leer: Ley del Impuesto sobre la Renta')),
      );
      await tester.pumpAndSettle();

      expect(harness.openedUrls, [
        Uri.parse('https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'),
      ]);
    });

    testWidgets('a source with no official link shows only its reference', (
      tester,
    ) async {
      await _open(harness, tester);
      final toggle = _inLine(4, find.text('Fundamento: Anexo 8 RMF 2026 B.VI'));
      await tester.ensureVisible(toggle);
      await tester.pumpAndSettle();

      await tester.tap(toggle);
      await tester.pumpAndSettle();

      expect(_inLine(4, find.text('Anexo 8 RMF 2026, B.VI')), findsOneWidget);
      expect(_inLine(4, find.textContaining('Leer')), findsNothing);
    });
  });

  testWidgets('the rounding adjustment is explained and cites nothing', (
    tester,
  ) async {
    await _open(
      harness,
      tester,
      reply: breakdownFrom(
        breakdownJson(lines: [...isrLines612.take(8), roundingLine(9)]),
      ),
    );
    await _reveal(tester, _line(9));

    expect(_inLine(9, find.text('Ajuste por redondeo')), findsOneWidget);
    expect(_inLine(9, find.text(r'$0.01')), findsOneWidget);
    expect(
      _inLine(9, find.textContaining('redondeada al centavo')),
      findsOneWidget,
    );
    expect(_inLine(9, find.textContaining('Fundamento')), findsNothing);
  });

  testWidgets('withholdings to remit are their own group, apart from the '
      'taxes', (tester) async {
    await _open(
      harness,
      tester,
      reply: breakdownFrom(
        breakdownJson(
          lines: [...isrLines612, ...ivaLines612(), ...withholdingLines],
        ),
      ),
    );
    await _reveal(tester, find.byKey(const ValueKey('group-withholding')));

    expect(
      _inGroup('withholding', find.text('Retenciones a enterar')),
      findsOneWidget,
    );
    expect(
      _inGroup('withholding', find.textContaining('Se paga aparte')),
      findsOneWidget,
    );
    expect(_inGroup('withholding', _line(13)), findsOneWidget);
    expect(_inGroup('isr', _line(13)), findsNothing);
    expect(_inGroup('iva', _line(13)), findsNothing);
  });

  testWidgets('IVA explains that it is one payment for every regime', (
    tester,
  ) async {
    await _open(harness, tester);
    await _reveal(tester, find.byKey(const ValueKey('group-iva')));

    expect(
      _inGroup('iva', find.textContaining('un solo pago')),
      findsOneWidget,
    );
    expect(
      _inGroup('iva', find.textContaining('aún no la valida un contador')),
      findsNothing,
    );
  });

  testWidgets('IVA netted across regimes carries the CPA warning (T2, T8)', (
    tester,
  ) async {
    await _open(
      harness,
      tester,
      reply: breakdownFrom(
        breakdownJson(
          lines: [...isrLines612, ...ivaLines612(consolidated: true)],
          disclaimers: [notTaxAdvice, cpaReviewDisclaimer],
          regime: regimeJson(
            '612',
            name: 'Actividades Empresariales y Profesionales',
            isrCents: 15096,
            cpaTopics: ['cross_regime_iva_netting'],
          ),
        ),
      ),
    );

    final banner = find.byKey(const ValueKey('cpa-review-banner'));
    expect(banner, findsOneWidget);
    expect(find.text('Requiere validación de un contador'), findsOneWidget);
    expect(
      tester.getBottomLeft(banner).dy,
      lessThan(tester.getTopLeft(_line(1)).dy),
      reason: 'the strong warning sits above the figures',
    );
    expect(
      find.textContaining('la compensación del IVA entre regímenes'),
      findsOneWidget,
    );
    await _reveal(tester, _line(12));
    expect(
      _inLine(12, find.text('Sin validar por un contador')),
      findsOneWidget,
    );
    expect(_inLine(1, find.text('Sin validar por un contador')), findsNothing);
    expect(
      _inGroup('iva', find.textContaining('aún no la valida un contador')),
      findsOneWidget,
    );
  });

  testWidgets('the legend is always on screen, above the lines', (
    tester,
  ) async {
    const serverLegend = 'Leyenda servida por el servidor.';
    await _open(
      harness,
      tester,
      reply: breakdownFrom(breakdownJson(legend: serverLegend)),
    );

    final legend = find.text(serverLegend);
    expect(legend, findsOneWidget);
    expect(
      tester.getTopLeft(legend).dy,
      lessThan(tester.getTopLeft(find.byKey(const ValueKey('group-isr'))).dy),
    );
    expect(
      find.bySemanticsLabel(RegExp(r'^Leyenda servida por el servidor\.$')),
      findsOneWidget,
    );
  });

  group('an incomplete regime', () {
    testWidgets('lists what is missing and marks the figures that depend '
        'on it', (tester) async {
      await _open(harness, tester, reply: _incomplete());

      expect(find.text('Falta:'), findsOneWidget);
      expect(
        find.text(
          'Pagos provisionales anteriores del ejercicio: enero de 2026, '
          'febrero de 2026',
        ),
        findsOneWidget,
      );
      expect(
        find.bySemanticsLabel(RegExp(r'^Base gravable \(incompleto\): ')),
        findsOneWidget,
      );
      expect(
        find.bySemanticsLabel(
          RegExp('^Ingresos cobrados acumulados ene–mar: '),
        ),
        findsOneWidget,
        reason: 'a figure from the records does not depend on the inputs',
      );
    });

    testWidgets('jumps to the inputs form of that regime and year', (
      tester,
    ) async {
      await _open(harness, tester, reply: _incomplete());

      await tester.tap(find.text('Completa tus datos'));
      await tester.pumpAndSettle();

      expect(
        harness.location,
        TaxEngineRoutes.taxInputs(
          fixtureTaxpayerId,
          '612',
          fiscalYear: 2026,
          field: 'prior_provisional_payments',
        ),
      );
    });

    testWidgets('after saving the inputs, comes back to this breakdown, '
        'read again', (tester) async {
      await _open(harness, tester, reply: _incomplete());
      await tester.tap(find.text('Completa tus datos'));
      await tester.pumpAndSettle();
      final reads = harness.repository.breakdownCalls.length;
      harness.repository.breakdownFallback = breakdownFrom(breakdownJson());

      await tester.tap(find.widgetWithText(FilledButton, 'Guardar'));
      await tester.pumpAndSettle();

      expect(harness.location, _breakdown612);
      expect(harness.repository.breakdownCalls.length, greaterThan(reads));
      expect(find.text('Falta:'), findsNothing);
    });
  });

  testWidgets('an informational regime explains why it has no lines', (
    tester,
  ) async {
    await _open(
      harness,
      tester,
      reply: breakdownFrom(
        breakdownJson(
          lines: const [],
          regime: regimeJson(
            '606',
            status: 'informational',
            treatment: 'informational',
            explanation: 'Pagas trimestralmente; no lo estimamos.',
            iva: null,
          ),
        ),
      ),
    );

    expect(
      find.text('Pagas trimestralmente; no lo estimamos.'),
      findsOneWidget,
    );
    expect(find.textContaining('no tiene desglose'), findsOneWidget);
    expect(find.byKey(const ValueKey('group-isr')), findsNothing);
  });

  group('loading and failures', () {
    testWidgets('announces the load while it is pending', (tester) async {
      final pending = Completer<Breakdown>();
      harness.repository.breakdownReplies.add(pending);
      await harness.pump(tester, location: null);
      unawaited(
        harness.container.read(appRouterProvider).push<void>(_breakdown612),
      );
      await tester.pump();
      await tester.pump();

      expect(find.text('Cargando el desglose…'), findsOneWidget);
      pending.complete(breakdownFrom(breakdownJson()));
      await tester.pumpAndSettle();
      expect(_line(1), findsOneWidget);
    });

    testWidgets('an unreachable server offers a retry', (tester) async {
      harness.repository.breakdownReplies.add(const ApiUnreachable());
      await _open(harness, tester);

      expect(find.textContaining('No hay conexión'), findsOneWidget);
      await tester.tap(find.text('Reintentar'));
      await tester.pumpAndSettle();

      expect(harness.repository.breakdownCalls, hasLength(2));
      expect(_line(1), findsOneWidget);
    });

    testWidgets('a regime not held that month says so', (tester) async {
      harness.repository.breakdownReplies.add(
        const ApiRejected(
          code: ErrorCode.businessRuleViolation,
          details: [
            FieldError(
              field: 'regime',
              code: 'regime_not_active',
              message: 'x',
            ),
          ],
        ),
      );
      await _open(harness, tester);

      expect(find.text('No tenías este régimen en ese mes.'), findsOneWidget);
    });
  });

  testWidgets('survives 200 % text on a phone', (tester) async {
    tester.view
      ..physicalSize = _phone
      ..devicePixelRatio = 1;
    addTearDown(tester.view.reset);
    tester.platformDispatcher.textScaleFactorTestValue = 2;
    addTearDown(tester.platformDispatcher.clearTextScaleFactorTestValue);

    await _open(
      harness,
      tester,
      reply: breakdownFrom(
        breakdownJson(
          lines: [...isrLines612, ...ivaLines612(), ...withholdingLines],
        ),
      ),
    );
    await tester.tap(find.text('Fundamento: LISR 102'));
    await tester.pumpAndSettle();
    await _reveal(tester, _line(13));

    expect(tester.takeException(), isNull);
  });

  for (final brightness in Brightness.values) {
    testWidgets('meets contrast and target guidelines (${brightness.name})', (
      tester,
    ) async {
      await harness.pump(
        tester,
        brightness: brightness,
        location: _breakdown612,
      );

      await expectLater(tester, meetsGuideline(textContrastGuideline));
      await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
      await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
    });
  }
}
