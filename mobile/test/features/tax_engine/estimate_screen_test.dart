import 'dart:async';

import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/api/generated/models/taxpayer_estimate.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_screen.dart';
import 'package:control_contable/features/tax_engine/tax_engine_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../taxpayers/taxpayers_fakes.dart';
import 'estimate_fixtures.dart';
import 'pump_estimate.dart';

/// Scrolls [finder] into view.
Future<void> _reveal(WidgetTester tester, Finder finder) async {
  await tester.scrollUntilVisible(
    finder,
    200,
    scrollable: find
        .descendant(
          of: find.byType(EstimateScreen),
          matching: find.byType(Scrollable),
        )
        .first,
  );
  await tester.pumpAndSettle();
}

/// A small phone in portrait (logical pixels).
const Size _phone = Size(360, 740);

Finder _inTotals(Finder finder) => find.descendant(
  of: find.byKey(const ValueKey('estimate-totals')),
  matching: finder,
);

Finder _inCard(String regimeCode, Finder finder) => find.descendant(
  of: find.byKey(ValueKey('regime-$regimeCode')),
  matching: finder,
);

void main() {
  late EstimateHarness harness;

  setUp(() => harness = EstimateHarness());

  group('a complete month (TX-612-TOTAL-01)', () {
    testWidgets('shows the totals, due date and the current month', (
      tester,
    ) async {
      await harness.pump(tester);

      expect(harness.repository.calls, [
        (taxpayerId: fixtureTaxpayerId, period: null),
      ]);
      expect(find.text('Estimado del mes'), findsOneWidget);
      expect(find.text('marzo de 2026'), findsWidgets);
      expect(find.textContaining('Mes en curso'), findsOneWidget);
      expect(_inTotals(find.text('ISR a pagar')), findsOneWidget);
      expect(_inTotals(find.text(r'$150.96')), findsOneWidget);
      expect(_inTotals(find.text('IVA a pagar')), findsOneWidget);
      expect(_inTotals(find.text(r'$2,133.33')), findsOneWidget);
      expect(find.text('Total a pagar'), findsOneWidget);
      expect(find.text('Vence el 17 de abril de 2026'), findsOneWidget);
      expect(find.text('Fecha sujeta a confirmación.'), findsOneWidget);
      expect(find.text('Retenciones a enterar'), findsNothing);
    });

    testWidgets('the total is read as pesos and centavos', (tester) async {
      await harness.pump(tester);

      expect(
        find.bySemanticsLabel(
          RegExp(r'^Total a pagar: 2,284 pesos con 29 centavos$'),
        ),
        findsOneWidget,
        reason: 'the total is its own screen-reader stop',
      );
      expect(
        find.bySemanticsLabel(
          RegExp(
            '^Esto es un aproximado. No todos los gastos aplican para tu '
            r'giro comercial\.$',
          ),
        ),
        findsOneWidget,
        reason: 'the legend is read on its own, never lost in a blob',
      );
      expect(
        find.bySemanticsLabel(RegExp(r'\$2,284\.29')),
        findsNothing,
        reason: 'the visual amount is excluded from semantics',
      );
    });

    testWidgets('renders the server legend right under the totals', (
      tester,
    ) async {
      const serverLegend = 'Leyenda servida por el servidor.';
      await harness.pump(
        tester,
        reply: estimateFrom({...estimateJson(), 'legend': serverLegend}),
      );

      final legend = find.text(serverLegend);
      expect(legend, findsOneWidget);
      final totalsBottom = tester
          .getBottomLeft(find.byKey(const ValueKey('estimate-totals')))
          .dy;
      final legendTop = tester.getTopLeft(legend).dy;
      expect(legendTop, greaterThan(totalsBottom));
      await _reveal(tester, find.byKey(const ValueKey('regime-612')));
      expect(
        tester.getTopLeft(legend).dy,
        lessThan(
          tester.getTopLeft(find.byKey(const ValueKey('regime-612'))).dy,
        ),
      );
    });

    testWidgets('shows the regime card as complete and every disclaimer', (
      tester,
    ) async {
      await harness.pump(tester);
      await _reveal(tester, find.byKey(const ValueKey('regime-612')));

      expect(
        _inCard('612', find.text('Actividades Empresariales y Profesionales')),
        findsOneWidget,
      );
      expect(_inCard('612', find.text('Completo')), findsOneWidget);
      expect(_inCard('612', find.text('Falta:')), findsNothing);
      await _reveal(tester, find.textContaining('Estimación informativa'));
      expect(find.textContaining('Estimación informativa'), findsOneWidget);
    });

    testWidgets('a regime links to its breakdown', (tester) async {
      await harness.pump(tester);
      await _reveal(tester, _inCard('612', find.text('Ver desglose')));

      await tester.tap(_inCard('612', find.text('Ver desglose')));
      await tester.pumpAndSettle();

      expect(
        harness.location,
        TaxEngineRoutes.breakdown(fixtureTaxpayerId, '2026-03', '612'),
      );
    });
  });

  group('an incomplete regime (TX-612-INC-01)', () {
    TaxpayerEstimate incomplete() => estimateFrom(
      estimateJson(
        disclaimers: [notTaxAdvice, incompleteInputsDisclaimer],
        isrCents: 415096,
        regimes: [
          regimeJson(
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
        ],
      ),
    );

    testWidgets('marks the figures as depending on the missing inputs', (
      tester,
    ) async {
      await harness.pump(tester, reply: incomplete());

      expect(
        find.descendant(
          of: find.byKey(const ValueKey('estimate-totals')),
          matching: find.text('Incompleto'),
        ),
        findsOneWidget,
      );
      await _reveal(tester, find.text('Completa tus datos'));
      expect(_inCard('612', find.text('Incompleto')), findsOneWidget);
      expect(
        _inCard(
          '612',
          find.text('Cifras con datos faltantes: se usaron ceros.'),
        ),
        findsOneWidget,
      );
      expect(_inCard('612', find.text('Falta:')), findsOneWidget);
      expect(
        _inCard(
          '612',
          find.text(
            'Pagos provisionales anteriores del ejercicio: enero de 2026, '
            'febrero de 2026',
          ),
        ),
        findsOneWidget,
      );
      expect(find.textContaining('Faltan datos para este cálculo'), findsOne);
    });

    testWidgets('no figure reads as final, on screen or to a screen reader', (
      tester,
    ) async {
      await harness.pump(tester, reply: incomplete());

      expect(_inTotals(find.text('Total a pagar (incompleto)')), findsOne);
      expect(_inTotals(find.text('ISR a pagar (incompleto)')), findsOne);
      expect(
        find.bySemanticsLabel(
          RegExp(
            r'^Total a pagar \(incompleto\): 6,284 pesos con 29 centavos$',
          ),
        ),
        findsOneWidget,
        reason: 'each amount is its own screen-reader stop, not one blob',
      );
      await _reveal(tester, find.text('Completa tus datos'));
      expect(_inCard('612', find.text('ISR a pagar (incompleto)')), findsOne);
      expect(
        find.bySemanticsLabel(RegExp('^ISR a pagar: ')),
        findsNothing,
        reason: 'an incomplete ISR is never announced as final',
      );
    });

    testWidgets('the CTA opens the tax inputs of that regime and year', (
      tester,
    ) async {
      await harness.pump(tester, reply: incomplete());
      await _reveal(tester, find.text('Completa tus datos'));

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
  });

  testWidgets('a partial regime shows its chip and what is not estimated', (
    tester,
  ) async {
    await harness.pump(
      tester,
      reply: estimateFrom(
        estimateJson(
          regimes: [
            regimeJson(
              '625',
              treatment: 'partial',
              status: 'partial',
              explanation: 'Calcula tu pago en 612 o RESICO.',
            ),
          ],
        ),
      ),
    );
    await _reveal(tester, find.byKey(const ValueKey('regime-625')));

    expect(_inCard('625', find.text('Parcial')), findsOneWidget);
    expect(
      _inCard('625', find.text('Calcula tu pago en 612 o RESICO.')),
      findsOneWidget,
    );
  });

  testWidgets('an informational regime explains itself with no amount', (
    tester,
  ) async {
    await harness.pump(
      tester,
      reply: estimateFrom(
        estimateJson(
          regimes: [
            regimeJson('612', isrCents: 15096),
            regimeJson(
              '605',
              name: 'Sueldos y salarios',
              treatment: 'informational',
              status: 'informational',
              iva: null,
              explanation: 'Tu patrón calcula y retiene tu ISR.',
            ),
          ],
        ),
      ),
    );
    await _reveal(tester, find.byKey(const ValueKey('regime-605')));

    expect(_inCard('605', find.text('Informativo')), findsOneWidget);
    expect(
      _inCard('605', find.text('Tu patrón calcula y retiene tu ISR.')),
      findsOneWidget,
    );
    expect(_inCard('605', find.textContaining(r'$')), findsNothing);
    expect(_inCard('605', find.text('Ver desglose')), findsNothing);
  });

  testWidgets('IVA in favor reads as saldo a favor, never as a payment', (
    tester,
  ) async {
    await harness.pump(
      tester,
      reply: estimateFrom(estimateJson(ivaCents: -50000)),
    );

    expect(find.text('IVA: saldo a favor'), findsOneWidget);
    expect(find.text(r'$500.00'), findsOneWidget);
    expect(find.text('IVA a pagar'), findsNothing);
    expect(
      find.bySemanticsLabel(RegExp('IVA: saldo a favor.*500 pesos')),
      findsOneWidget,
    );
  });

  testWidgets('withholdings to remit show apart from the taxes '
      '(TX-601-REMIT-01)', (tester) async {
    await harness.pump(
      tester,
      reply: estimateFrom(
        estimateJson(
          ivaCents: 12000000,
          isrCents: 3600000,
          withholdingsToRemitCents: 1033333,
          regimes: [
            regimeJson('601', isrCents: 3600000, withheldToRemitCents: 1033333),
          ],
        ),
      ),
    );

    expect(
      find.descendant(
        of: find.byKey(const ValueKey('estimate-totals')),
        matching: find.text('Retenciones a enterar'),
      ),
      findsOneWidget,
    );
    expect(find.text(r'$10,333.33'), findsWidgets);
    expect(find.text(r'$166,333.33'), findsOneWidget);
  });

  testWidgets('the strong CPA warning sits above the figures and names the '
      'topics on the regime (TX-NET-01)', (tester) async {
    await harness.pump(
      tester,
      reply: estimateFrom(
        estimateJson(
          ivaCents: 20000,
          disclaimers: [notTaxAdvice, cpaReviewDisclaimer],
          regimes: [
            regimeJson('612', isrCents: 15096),
            regimeJson(
              '606',
              isrCents: 0,
              iva: {
                'charged_collected_cents': 320000,
                'creditable_paid_cents': 300000,
                'withheld_from_taxpayer_cents': 213333,
                'definitive_excluded_cents': 0,
              },
              cpaTopics: ['cross_regime_iva_netting'],
            ),
          ],
        ),
      ),
    );

    final banner = find.byKey(const ValueKey('cpa-review-banner'));
    expect(banner, findsOneWidget);
    expect(
      find.descendant(
        of: banner,
        matching: find.textContaining('un contador aún no valida'),
      ),
      findsOneWidget,
    );
    expect(
      tester.getBottomLeft(banner).dy,
      lessThanOrEqualTo(
        tester.getTopLeft(find.byKey(const ValueKey('estimate-totals'))).dy,
      ),
    );
    expect(
      find.text('IVA consolidado de 2 regímenes: se declara uno solo por RFC.'),
      findsOneWidget,
    );
    await _reveal(tester, find.byKey(const ValueKey('regime-606')));
    expect(
      _inCard(
        '606',
        find.text(
          'Pendiente de validar por un contador: la compensación del IVA '
          'entre regímenes.',
        ),
      ),
      findsOneWidget,
    );
  });

  testWidgets('a provisional-parameters warning is shown as a warning', (
    tester,
  ) async {
    await harness.pump(
      tester,
      reply: estimateFrom(
        estimateJson(
          disclaimers: [notTaxAdvice, provisionalParametersDisclaimer],
        ),
      ),
    );

    final warning = find.byKey(
      const ValueKey('disclaimer-provisional_parameters'),
    );
    await _reveal(tester, warning);
    expect(
      find.descendant(
        of: warning,
        matching: find.text(
          'Aún no se publican las tarifas de 2027; usamos las de 2026.',
        ),
      ),
      findsOneWidget,
    );
    expect(find.byKey(const ValueKey('cpa-review-banner')), findsNothing);
  });

  group('loading and failures', () {
    testWidgets('a skeleton shows while the estimate loads', (tester) async {
      final pending = Completer<TaxpayerEstimate>();
      harness.repository.replies.add(pending);
      await harness.pump(tester, location: null);
      unawaited(
        harness.container
            .read(appRouterProvider)
            .push<void>(TaxEngineRoutes.estimate),
      );
      await tester.pump();
      await tester.pump();

      expect(find.text('Cargando el estimado…'), findsOneWidget);
      expect(find.byKey(const ValueKey('estimate-skeleton')), findsOneWidget);

      pending.complete(estimateFrom(completeEstimateJson));
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('estimate-skeleton')), findsNothing);
      expect(find.text('Total a pagar'), findsOneWidget);
    });

    testWidgets('an unreachable server offers a retry', (tester) async {
      harness.repository.replies.add(const ApiUnreachable());
      await harness.pump(tester);

      expect(
        find.text('No hay conexión. Revisa tu internet e inténtalo de nuevo.'),
        findsOneWidget,
      );
      await tester.tap(find.text('Reintentar'));
      await tester.pumpAndSettle();

      expect(harness.repository.calls, hasLength(2));
      expect(find.text('Total a pagar'), findsOneWidget);
    });

    testWidgets('a taxpayer gone from the server says so', (tester) async {
      harness.repository.replies.add(const ApiNotFound());
      await harness.pump(tester);

      expect(
        find.text(
          'No se encontró este contribuyente. Puede haberse eliminado en otro '
          'dispositivo.',
        ),
        findsOneWidget,
      );
    });

    testWidgets('a month without parameters explains why', (tester) async {
      harness.repository.replies.add(
        ApiRejected(
          code: ErrorCode.businessRuleViolation,
          details: [detail('period', 'parameters_unavailable')],
        ),
      );
      await harness.pump(tester);

      expect(
        find.text(
          'No hay tarifas cargadas para ese mes. Elige un mes desde enero de '
          '2026.',
        ),
        findsOneWidget,
      );
    });

    testWidgets('a future month refused by the server explains why', (
      tester,
    ) async {
      harness.repository.replies.add(
        ApiValidationFailure([detail('period', 'out_of_range')]),
      );
      await harness.pump(tester);

      expect(
        find.text(
          'Ese mes todavía no llega. Elige el mes en curso o uno anterior.',
        ),
        findsOneWidget,
      );
    });

    testWidgets('a taxpayer with no regime in the month has nothing to '
        'estimate', (tester) async {
      await harness.pump(
        tester,
        reply: estimateFrom(
          estimateJson(ivaCents: 0, isrCents: 0, regimes: const []),
        ),
      );

      expect(
        find.text(
          'Este contribuyente no tiene regímenes en marzo de 2026. No hay '
          'nada que estimar.',
        ),
        findsOneWidget,
      );
      expect(find.text('Total a pagar'), findsNothing);
    });
  });

  group('choosing the month', () {
    testWidgets('lists the current month back to January 2026 and explains '
        'the limits', (tester) async {
      await harness.pump(tester);

      expect(
        find.text(
          'Se puede estimar desde enero de 2026 hasta el mes en curso. Los '
          'meses futuros aún no se pueden estimar y antes de 2026 no hay '
          'tarifas cargadas.',
        ),
        findsOneWidget,
      );
      await tester.tap(find.byKey(const ValueKey('estimate-period')));
      await tester.pumpAndSettle();

      expect(find.text('enero de 2026').last, findsOneWidget);
      expect(find.text('abril de 2026'), findsNothing);
      expect(find.text('diciembre de 2025'), findsNothing);
    });

    testWidgets('a past month is estimated again from the server', (
      tester,
    ) async {
      await harness.pump(tester);
      harness.repository.fallback = estimateFrom(
        estimateJson(period: '2026-02', dueDate: '2026-03-17'),
      );

      await tester.tap(find.byKey(const ValueKey('estimate-period')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('febrero de 2026').last);
      await tester.pumpAndSettle();

      expect(harness.repository.calls.last, (
        taxpayerId: fixtureTaxpayerId,
        period: '2026-02',
      ));
      expect(find.textContaining('Mes en curso'), findsNothing);
      expect(find.text('Vence el 17 de marzo de 2026'), findsOneWidget);
    });

    testWidgets('pulling down refreshes the estimate', (tester) async {
      await harness.pump(tester);

      await tester.fling(
        find.text('Total a pagar'),
        const Offset(0, 400),
        1000,
      );
      await tester.pumpAndSettle();

      expect(harness.repository.calls, hasLength(2));
    });
  });

  testWidgets('survives 200 % text on a phone', (tester) async {
    tester.view
      ..physicalSize = _phone
      ..devicePixelRatio = 1;
    addTearDown(tester.view.reset);
    tester.platformDispatcher.textScaleFactorTestValue = 2;
    addTearDown(tester.platformDispatcher.clearTextScaleFactorTestValue);

    await harness.pump(
      tester,
      reply: estimateFrom(
        estimateJson(
          ivaCents: 12000000,
          isrCents: 3600000,
          withholdingsToRemitCents: 1033333,
        ),
      ),
    );
    await _reveal(tester, find.text('Ver desglose'));

    expect(tester.takeException(), isNull);
  });

  for (final brightness in Brightness.values) {
    testWidgets('meets contrast and target guidelines (${brightness.name})', (
      tester,
    ) async {
      await harness.pump(
        tester,
        brightness: brightness,
        reply: estimateFrom(
          estimateJson(
            disclaimers: [notTaxAdvice, cpaReviewDisclaimer],
            regimes: [
              regimeJson(
                '612',
                status: 'incomplete',
                isrCents: 15096,
                missingInputs: [
                  {'key': 'ptu_paid', 'label': 'PTU pagada'},
                ],
              ),
            ],
          ),
        ),
      );

      await expectLater(tester, meetsGuideline(textContrastGuideline));
      await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
      await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
    });
  }
}
