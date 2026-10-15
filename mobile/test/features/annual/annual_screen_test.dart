import 'dart:async';

import 'package:control_contable/api/generated/models/annual_computed_section_kind.dart';
import 'package:control_contable/api/generated/models/annual_estimate.dart';
import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/features/annual/annual_routes.dart';
import 'package:control_contable/features/annual/presentation/annual_screen.dart';
import 'package:control_contable/features/tax_engine/tax_engine_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../tax_engine/estimate_fixtures.dart' show fixtureTaxpayerId;
import '../taxpayers/taxpayers_fakes.dart' show detail;
import 'annual_fixtures.dart';
import 'annual_inputs_fixtures.dart';
import 'pump_annual.dart';

/// AN-06's suggested coeficiente, labelled as MODULE-annual §2.7 asks.
const String _suggestedCu =
    'Coeficiente de utilidad sugerido para tus pagos de 2027: 0.1380. '
    'Confírmalo con tu contador.';

/// A small phone in portrait (logical pixels).
const Size _phone = Size(360, 740);

Finder _section(String key) => find.byKey(ValueKey('annual-section-$key'));

Finder _inSection(String key, Finder finder) =>
    find.descendant(of: _section(key), matching: finder);

final Finder _headline = find.byKey(const ValueKey('annual-headline'));

Finder _inHeadline(Finder finder) =>
    find.descendant(of: _headline, matching: finder);

Future<void> _reveal(WidgetTester tester, Finder finder) async {
  await tester.scrollUntilVisible(
    finder,
    200,
    scrollable: find
        .descendant(
          of: find.byType(AnnualScreen),
          matching: find.byType(Scrollable),
        )
        .first,
  );
  await tester.pumpAndSettle();
}

AnnualEstimate _estimate({
  List<Map<String, Object?>>? sections,
  List<String> headline = const ['pf_art152'],
  List<Map<String, Object?>>? disclaimers,
  String personType = 'fisica',
  String yearStatus = 'closed',
  String? deadline = '2027-04-30',
}) => annualEstimateFrom(
  annualEstimateJson(
    sections: sections,
    headline: headline,
    disclaimers: disclaimers,
    personType: personType,
    yearStatus: yearStatus,
    deadline: deadline,
  ),
);

void main() {
  late AnnualHarness harness;

  setUp(() => harness = AnnualHarness());

  group('year', () {
    testWidgets('October 2026 asks for 2026, the year in progress', (
      tester,
    ) async {
      await harness.pump(tester);

      expect(harness.repository.calls, [
        (taxpayerId: fixtureTaxpayerId, fiscalYear: 2026),
      ]);
      expect(find.text('Declaración anual'), findsOneWidget);
      expect(find.text('Ejercicio'), findsOneWidget);
    });

    testWidgets('until 30 April the previous year is shown; the picker '
        'offers the others', (tester) async {
      harness.now = DateTime.utc(2027, 4, 30, 18);
      await harness.pump(tester);

      expect(harness.repository.calls.single.fiscalYear, 2026);

      await tester.tap(find.byKey(const ValueKey('annual-year')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('2027').last);
      await tester.pumpAndSettle();

      expect(harness.repository.calls.last.fiscalYear, 2027);
    });

    testWidgets('pulling down reads the estimate again', (tester) async {
      await harness.pump(tester);

      await tester.fling(
        find.byType(ListView).first,
        const Offset(0, 400),
        1000,
      );
      await tester.pumpAndSettle();

      expect(harness.repository.calls, hasLength(2));
    });
  });

  group('headline', () {
    testWidgets('AN-01: saldo a favor, its deadline and one spoken summary', (
      tester,
    ) async {
      await harness.pump(tester);

      expect(_inHeadline(find.text('Saldo a favor')), findsOneWidget);
      expect(
        _inHeadline(find.text(r'$9,408.00', findRichText: true)),
        findsOneWidget,
      );
      expect(
        _inHeadline(find.text('Vence el 30 de abril de 2027')),
        findsOneWidget,
      );
      expect(
        find.bySemanticsLabel(
          'Declaración anual 2026, estimación. Saldo a favor, 9,408 pesos. '
          'Vence el 30 de abril de 2027.',
        ),
        findsOneWidget,
      );
    });

    testWidgets('the legend is always next to the figures, from the server', (
      tester,
    ) async {
      await harness.pump(tester);

      final legend = find.text(annualLegendText);
      expect(legend, findsOneWidget);
      final gap =
          tester.getTopLeft(_headline).dy - tester.getBottomLeft(legend).dy;
      expect(gap, inInclusiveRange(0, 40));
    });

    testWidgets('a saldo a cargo reads in words; incomplete and the year in '
        'progress are spoken too', (tester) async {
      await harness.pump(
        tester,
        reply: _estimate(
          yearStatus: 'in_progress',
          sections: [
            sectionJson(
              status: 'incomplete',
              tax: annualTaxJson(balanceKind: 'a_cargo', balanceCents: 19987),
              missingInputs: [
                missingJson(
                  'salary_statement',
                  'Constancia de sueldos',
                  'salary_statements',
                ),
              ],
            ),
          ],
        ),
      );

      expect(_inHeadline(find.text('Saldo a cargo')), findsOneWidget);
      expect(_inHeadline(find.text('Incompleto')), findsOneWidget);
      expect(_inHeadline(find.text('Año en curso')), findsOneWidget);
      expect(
        find.bySemanticsLabel(
          'Declaración anual 2026, estimación. Saldo a cargo, 199 pesos con '
          '87 centavos. Vence el 30 de abril de 2027. Incompleto. Año en '
          'curso.',
        ),
        findsOneWidget,
      );
    });

    testWidgets('a zero balance reads "Sin saldo"', (tester) async {
      await harness.pump(
        tester,
        reply: _estimate(
          sections: [
            sectionJson(
              tax: annualTaxJson(balanceKind: 'cero', balanceCents: 0),
            ),
          ],
        ),
      );

      expect(_inHeadline(find.text('Sin saldo')), findsOneWidget);
    });

    testWidgets('"Ver cómo se calculó" opens the section breakdown', (
      tester,
    ) async {
      await harness.pump(tester);

      await tester.tap(_inHeadline(find.text('Ver cómo se calculó')));
      await tester.pumpAndSettle();

      expect(
        harness.location,
        AnnualRoutes.breakdown(
          fixtureTaxpayerId,
          2026,
          AnnualComputedSectionKind.pfArt152,
        ),
      );
    });

    testWidgets('no headline section: no balance is invented', (tester) async {
      await harness.pump(
        tester,
        reply: _estimate(
          headline: [],
          deadline: null,
          sections: [resicoSectionJson()],
          disclaimers: [annualEstimateDisclaimer, annualCpaDisclaimer],
        ),
      );

      expect(
        _inHeadline(find.text('No hay saldo que estimar este año.')),
        findsOneWidget,
      );
      expect(find.textContaining('Saldo a'), findsNothing);
    });
  });

  group('sections', () {
    testWidgets('PF art. 152: figures, credits with their sources, '
        'obligation and notes', (tester) async {
      await harness.pump(
        tester,
        reply: _estimate(
          sections: [
            sectionJson(
              tax: annualTaxJson(
                salaryIsrCents: 500000,
                filedMonths: 6,
                estimatedMonths: 3,
              ),
              notes: [
                {
                  'code': 'refund_only_paid',
                  'text':
                      'Solo se devuelve lo efectivamente pagado o retenido.',
                },
              ],
            ),
          ],
        ),
      );
      await _reveal(tester, _section('pf_art152'));

      for (final label in [
        'Ingresos acumulables',
        'Deducciones personales',
        'Base gravable',
        'ISR del ejercicio',
        'Pagos provisionales',
        'Retenciones de ISR',
        'ISR retenido por sueldos',
      ]) {
        expect(_inSection('pf_art152', find.text(label)), findsOneWidget);
      }
      expect(
        _inSection('pf_art152', find.text(r'$500,000.00')),
        findsOneWidget,
      );
      expect(
        _inSection(
          'pf_art152',
          find.text('Pagos: 6 meses confirmados, 3 estimados por la app.'),
        ),
        findsOneWidget,
      );
      expect(
        _inSection('pf_art152', find.text('Estás obligado a presentarla.')),
        findsOneWidget,
      );
      expect(
        _inSection(
          'pf_art152',
          find.text(
            'Tuviste ingresos por actividad empresarial o profesional.',
          ),
        ),
        findsOneWidget,
      );
      expect(
        _inSection(
          'pf_art152',
          find.text('Solo se devuelve lo efectivamente pagado o retenido.'),
        ),
        findsOneWidget,
      );
    });

    testWidgets('RESICO PF is an informative check, never a balance', (
      tester,
    ) async {
      await harness.pump(
        tester,
        reply: _estimate(
          headline: [],
          sections: [resicoSectionJson()],
          disclaimers: [annualEstimateDisclaimer, annualCpaDisclaimer],
        ),
      );
      await _reveal(tester, _section('pf_resico_check'));

      expect(
        _inSection('pf_resico_check', find.text('Informativo')),
        findsOneWidget,
      );
      expect(
        _inSection('pf_resico_check', find.text('Diferencia informativa')),
        findsOneWidget,
      );
      expect(
        _inSection('pf_resico_check', find.text(r'−$1,320.00')),
        findsOneWidget,
      );
      expect(
        _inSection('pf_resico_check', find.text('Tasa anual: 1.1\u00a0%')),
        findsOneWidget,
      );
      expect(
        _inSection(
          'pf_resico_check',
          find.text(
            'No es un saldo a cargo ni a favor: no se paga ni se devuelve.',
          ),
        ),
        findsOneWidget,
      );
      expect(find.textContaining('Saldo a'), findsNothing);
    });

    testWidgets('PM 601: balance plus the informative PTU and suggested '
        'coeficiente, marked as unverified', (tester) async {
      harness.personType = PersonType.moral;
      await harness.pump(
        tester,
        reply: _estimate(
          personType: 'moral',
          headline: ['pm_601'],
          deadline: '2027-03-31',
          sections: [pm601SectionJson()],
        ),
      );

      expect(_inHeadline(find.text('Saldo a cargo')), findsOneWidget);
      expect(
        _inHeadline(find.text('Vence el 31 de marzo de 2027')),
        findsOneWidget,
      );
      await _reveal(tester, find.byKey(const ValueKey('pm-extras')));
      final extras = find.byKey(const ValueKey('pm-extras'));
      for (final text in [
        'Datos informativos',
        'Base de PTU',
        'PTU a repartir',
        r'$150,000.00',
        _suggestedCu,
        'Sin validar por un contador',
      ]) {
        expect(
          find.descendant(of: extras, matching: find.text(text)),
          findsOneWidget,
          reason: text,
        );
      }
      expect(
        _inSection('pm_601', find.text('Resultado fiscal')),
        findsOneWidget,
      );
      expect(
        _inSection('pm_601', find.text('Deducciones personales')),
        findsNothing,
      );
    });

    testWidgets('an explanation-only regime shows its text and deadline, no '
        'figure', (tester) async {
      harness.personType = PersonType.moral;
      await harness.pump(
        tester,
        reply: _estimate(
          personType: 'moral',
          headline: [],
          deadline: '2027-02-15',
          sections: [explanationSectionJson()],
        ),
      );
      await _reveal(tester, _section('explanation-603'));

      expect(
        _inSection('explanation-603', find.text('Régimen 603')),
        findsOneWidget,
      );
      expect(
        _inSection(
          'explanation-603',
          find.textContaining('tu contador la prepara'),
        ),
        findsOneWidget,
      );
      expect(
        _inSection(
          'explanation-603',
          find.text('Fecha límite: 15 de febrero de 2027'),
        ),
        findsOneWidget,
      );
      expect(
        _inSection('explanation-603', find.textContaining(r'$')),
        findsNothing,
      );
    });
  });

  group('incomplete', () {
    final missing = [
      missingJson(
        'salary_statement',
        'Constancia de sueldos',
        'salary_statements',
      ),
      missingJson(
        'prior_year_accumulable_income',
        'Ingresos acumulables del año anterior',
        'annual_inputs',
      ),
      missingJson(
        'filed_payment',
        'Pagos provisionales presentados',
        'filed_payments',
        regimeCode: '612',
        months: ['2026-03'],
      ),
      missingJson('ptu_paid', 'PTU pagada', 'tax_inputs', regimeCode: '612'),
      missingJson('donation', 'Donativos', 'personal_deductions'),
    ];

    Future<void> pumpIncomplete(WidgetTester tester) => harness.pump(
      tester,
      reply: _estimate(
        disclaimers: [annualEstimateDisclaimer, annualIncompleteDisclaimer],
        sections: [sectionJson(status: 'incomplete', missingInputs: missing)],
      ),
    );

    // Each missing input opens the screen that completes it (ANN-19..21).
    for (final (label, location) in [
      (
        'Constancia de sueldos',
        AnnualRoutes.salaryStatements(fixtureTaxpayerId, 2026),
      ),
      (
        'Ingresos acumulables del año anterior',
        AnnualRoutes.annualInputs(fixtureTaxpayerId, 2026),
      ),
      (
        'Pagos provisionales presentados',
        AnnualRoutes.filedPayments(fixtureTaxpayerId, 2026, '612'),
      ),
      ('Donativos', AnnualRoutes.personalDeductions(fixtureTaxpayerId, 2026)),
    ]) {
      testWidgets('"Falta: $label" opens its screen', (tester) async {
        await pumpIncomplete(tester);
        final cta = _inSection('pf_art152', find.text(label));
        await _reveal(tester, cta);

        expect(_inSection('pf_art152', find.text('Falta:')), findsOneWidget);
        await tester.tap(cta);
        await tester.pumpAndSettle();
        expect(harness.location, location);
      });
    }

    testWidgets('"Falta: PTU pagada" opens the tax-inputs form at its field', (
      tester,
    ) async {
      await pumpIncomplete(tester);
      final cta = _inSection('pf_art152', find.text('PTU pagada'));
      await _reveal(tester, cta);

      await tester.tap(cta);
      await tester.pumpAndSettle();

      expect(
        harness.location,
        TaxEngineRoutes.taxInputs(
          fixtureTaxpayerId,
          '612',
          fiscalYear: 2026,
          field: 'ptu_paid',
        ),
      );
    });

    testWidgets('the months of a missing payment are named', (tester) async {
      await harness.pump(
        tester,
        reply: _estimate(
          sections: [
            sectionJson(status: 'incomplete', missingInputs: [missing[2]]),
          ],
        ),
      );
      await _reveal(tester, _section('pf_art152'));

      expect(
        _inSection('pf_art152', find.text('marzo de 2026')),
        findsOneWidget,
      );
    });
  });

  group('datos para tu anual', () {
    testWidgets('PF: deductions and constancias with counters, payments per '
        'regime, the year facts', (tester) async {
      harness.inputs
        ..deductions.addAll([
          deductionJson(),
          deductionJson(id: 'ded-2'),
          deductionJson(id: 'ded-3'),
        ])
        ..statements.add(statementJson());
      await harness.pump(tester);
      final block = find.byKey(const ValueKey('annual-inputs'));
      await _reveal(tester, block);

      Finder inBlock(String text) =>
          find.descendant(of: block, matching: find.text(text));
      expect(inBlock('Datos para tu anual'), findsOneWidget);
      expect(inBlock('Deducciones personales'), findsOneWidget);
      expect(inBlock('3 registradas'), findsOneWidget);
      expect(inBlock('Constancias de sueldos'), findsOneWidget);
      expect(inBlock('1 registrada'), findsOneWidget);
      expect(inBlock('Pagos provisionales presentados'), findsOneWidget);
      expect(inBlock('Régimen 612'), findsOneWidget);
      expect(inBlock('Datos del año'), findsOneWidget);

      await tester.tap(inBlock('Deducciones personales'));
      await tester.pumpAndSettle();
      expect(
        harness.location,
        AnnualRoutes.personalDeductions(fixtureTaxpayerId, 2026),
      );
    });

    testWidgets('PM: only the provisional payments', (tester) async {
      harness.personType = PersonType.moral;
      await harness.pump(
        tester,
        reply: _estimate(
          personType: 'moral',
          headline: ['pm_601'],
          sections: [pm601SectionJson()],
        ),
      );
      final block = find.byKey(const ValueKey('annual-inputs'));
      await _reveal(tester, block);

      expect(
        find.descendant(
          of: block,
          matching: find.text('Deducciones personales'),
        ),
        findsNothing,
      );
      expect(
        find.descendant(of: block, matching: find.text('Régimen 601')),
        findsOneWidget,
      );
    });

    testWidgets('counters that fail to load are left out', (tester) async {
      harness.inputs.failures
        ..['listDeductions'] = const ApiUnreachable()
        ..['listStatements'] = const ApiUnreachable();
      await harness.pump(tester);
      final block = find.byKey(const ValueKey('annual-inputs'));
      await _reveal(tester, block);

      expect(
        find.descendant(
          of: block,
          matching: find.text('Deducciones personales'),
        ),
        findsOneWidget,
      );
      expect(find.textContaining('registrada'), findsNothing);
    });

    testWidgets('read-only: says what still works', (tester) async {
      harness.readOnly = true;
      await harness.pump(tester);
      final notice = find.byKey(const ValueKey('annual-read-only'));
      await _reveal(tester, notice);

      expect(
        find.descendant(
          of: notice,
          matching: find.text(
            'Tu periodo de prueba terminó: puedes consultar esta estimación, '
            'pero capturar o cambiar datos requiere un plan.',
          ),
        ),
        findsOneWidget,
      );
    });
  });

  group('disclaimers', () {
    testWidgets('the strong CPA banner comes before every figure; warnings '
        'and notes are shown', (tester) async {
      await harness.pump(
        tester,
        reply: _estimate(
          disclaimers: [
            annualEstimateDisclaimer,
            annualCpaDisclaimer,
            paymentsEstimatedDisclaimer,
            yearInProgressDisclaimer,
            noAnnualIvaDisclaimer,
          ],
          sections: [
            sectionJson(cpaTopics: ['uma_for_minimum_wage']),
          ],
        ),
      );

      final banner = find.byKey(const ValueKey('cpa-review-banner'));
      expect(banner, findsOneWidget);
      expect(
        tester.getBottomLeft(banner).dy,
        lessThanOrEqualTo(tester.getTopLeft(_headline).dy),
      );
      for (final code in [
        'payments_estimated',
        'year_in_progress',
        'no_annual_iva',
      ]) {
        await _reveal(tester, find.byKey(ValueKey('disclaimer-$code')));
      }
      await _reveal(tester, _section('pf_art152'));
      expect(
        _inSection(
          'pf_art152',
          find.text(
            'Pendiente de validar por un contador: el uso de la UMA en lugar '
            'del salario mínimo.',
          ),
        ),
        findsOneWidget,
      );
    });
  });

  group('loading and failures', () {
    testWidgets('a status line shows while the estimate loads', (tester) async {
      final pending = Completer<AnnualEstimate>();
      harness.repository.replies.add(pending);
      await harness.pump(tester, location: null);
      unawaited(
        harness.container
            .read(appRouterProvider)
            .push<void>(AnnualRoutes.estimate),
      );
      await tester.pump();
      await tester.pump();

      expect(find.text('Cargando la estimación anual…'), findsOneWidget);

      pending.complete(_estimate());
      await tester.pumpAndSettle();
      expect(find.text('Cargando la estimación anual…'), findsNothing);
    });

    final failures = <String, Object>{
      'No hay conexión. Revisa tu internet e inténtalo de nuevo.':
          const ApiUnreachable(),
      'Ese año todavía no llega. Elige el año en curso o uno anterior.':
          ApiValidationFailure([detail('fiscalYear', 'out_of_range')]),
      'No hay tarifas cargadas para ese año. Elige un año desde 2026.':
          ApiRejected(
            code: ErrorCode.businessRuleViolation,
            details: [detail('fiscalYear', 'parameters_unavailable')],
          ),
      'No se pudo cargar la estimación anual.': const ApiUnexpected(),
    };
    for (final MapEntry(key: message, value: failure) in failures.entries) {
      testWidgets('failure: $message', (tester) async {
        harness.repository.replies.add(failure);
        await harness.pump(tester);

        expect(find.text(message), findsOneWidget);
        await tester.tap(find.text('Reintentar'));
        await tester.pumpAndSettle();

        expect(harness.repository.calls, hasLength(2));
        expect(find.text('Saldo a favor'), findsWidgets);
      });
    }

    testWidgets('no section: nothing to estimate', (tester) async {
      await harness.pump(
        tester,
        reply: _estimate(headline: [], deadline: null, sections: []),
      );

      expect(
        find.text(
          'Este contribuyente no tiene regímenes en 2026: no hay '
          'declaración anual que estimar.',
        ),
        findsOneWidget,
      );
      expect(_headline, findsNothing);
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
      reply: _estimate(
        yearStatus: 'in_progress',
        sections: [
          sectionJson(
            status: 'incomplete',
            missingInputs: [
              missingJson(
                'salary_statement',
                'Constancia de sueldos',
                'salary_statements',
              ),
            ],
          ),
          resicoSectionJson(),
        ],
      ),
    );
    await _reveal(
      tester,
      find.byKey(const ValueKey('disclaimer-no_annual_iva')),
    );

    expect(tester.takeException(), isNull);
  });

  for (final brightness in Brightness.values) {
    testWidgets('meets contrast and target guidelines (${brightness.name})', (
      tester,
    ) async {
      await harness.pump(
        tester,
        brightness: brightness,
        reply: _estimate(
          yearStatus: 'in_progress',
          disclaimers: [
            annualEstimateDisclaimer,
            annualCpaDisclaimer,
            paymentsEstimatedDisclaimer,
          ],
          sections: [
            sectionJson(
              status: 'incomplete',
              missingInputs: [
                missingJson(
                  'salary_statement',
                  'Constancia de sueldos',
                  'salary_statements',
                ),
              ],
            ),
            resicoSectionJson(),
          ],
        ),
      );

      await expectLater(tester, meetsGuideline(textContrastGuideline));
      await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
      await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
    });
  }
}
