import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/annual/annual_routes.dart';
import 'package:control_contable/features/tax_engine/tax_engine_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../tax_engine/estimate_fixtures.dart' show fixtureTaxpayerId;
import '../taxpayers/pump_onboarding.dart' show inputLabeled, typeInto;
import 'annual_inputs_fixtures.dart';
import 'pump_annual_inputs.dart';

final String _screen = AnnualRoutes.annualInputs(fixtureTaxpayerId, 2026);

const String _priorYear = 'Ingresos acumulables de 2025 (opcional)';

/// ANN-21: the per-year facts (prior-year income, other exempt income,
/// told the employer) and, read-only with a link each, the regimes' annual
/// tax inputs (losses, PTU paid, depreciation).
void main() {
  late AnnualInputsHarness harness;

  setUp(() => harness = AnnualInputsHarness());

  Future<void> tapText(WidgetTester tester, String text) async {
    final target = find.text(text).last;
    await tester.ensureVisible(target);
    await tester.pumpAndSettle();
    await tester.tap(target);
    await tester.pumpAndSettle();
  }

  testWidgets('PF: the facts start at their defaults; the regime inputs are '
      'shown with their values', (tester) async {
    await harness.pump(tester, location: _screen);

    expect(find.text('Datos del año'), findsOneWidget);
    expect(
      tester.widget<TextField>(inputLabeled(_priorYear)).controller!.text,
      '',
    );
    expect(
      tester
          .widget<TextField>(inputLabeled('Otros ingresos exentos'))
          .controller!
          .text,
      '0.00',
    );
    expect(find.text('Datos de tus regímenes'), findsOneWidget);
    expect(find.text('PTU pagada'), findsOneWidget);
    expect(find.text(r'$12,000.00'), findsOneWidget);
    expect(find.text('Sin capturar'), findsNWidgets(2));
  });

  testWidgets('saves the facts; empty prior-year income is unknown (null)', (
    tester,
  ) async {
    await harness.pump(tester, location: _screen);
    await typeInto(tester, 'Otros ingresos exentos', '1,500');
    await tapText(
      tester,
      'Le avisé a mi empleador que presentaré mi declaración anual',
    );
    await tapText(tester, 'Guardar');

    final sent = harness.inputs.putFactsBodies.single;
    expect(sent.priorYearAccumulableIncomeCents, isNull);
    expect(sent.otherExemptIncomeCents, 150000);
    expect(sent.toldEmployerWillFile, isTrue);
    expect(find.text('Datos del año guardados'), findsOneWidget);
  });

  testWidgets('a typed prior-year income is sent in centavos', (tester) async {
    harness.inputs.facts = annualFactsOf(priorYearCents: 100);
    await harness.pump(tester, location: _screen);
    await typeInto(tester, _priorYear, '350,000');
    await tapText(tester, 'Guardar');

    expect(
      harness.inputs.putFactsBodies.single.priorYearAccumulableIncomeCents,
      35000000,
    );
  });

  testWidgets('each regime input opens its tax-inputs form at that key', (
    tester,
  ) async {
    await harness.pump(tester, location: _screen);
    final edit = find.byTooltip('Editar PTU pagada, régimen 612');
    await tester.ensureVisible(edit);
    await tester.pumpAndSettle();
    await tester.tap(edit);
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

  testWidgets('PM: only the regime inputs, nothing to save', (tester) async {
    harness.personType = PersonType.moral;
    await harness.pump(tester, location: _screen);

    expect(
      find.text(
        'Las personas morales sólo capturan los datos de sus regímenes.',
      ),
      findsOneWidget,
    );
    expect(find.text(_priorYear), findsNothing);
    expect(find.widgetWithText(FilledButton, 'Guardar'), findsNothing);
    expect(find.text('PTU pagada'), findsOneWidget);
  });

  testWidgets('a 402 opens the paywall', (tester) async {
    harness.inputs.failures['putFacts'] = const ApiEntitlementRequired();
    await harness.pump(tester, location: _screen);
    await tapText(tester, 'Guardar');

    expect(find.text('Tu cuenta es de solo consulta'), findsOneWidget);
  });

  testWidgets('a failed load offers a retry', (tester) async {
    harness.inputs.failures['getFacts'] = const ApiUnreachable();
    await harness.pump(tester, location: _screen);

    expect(find.text('No se pudieron cargar los datos del año.'), findsOne);
    harness.inputs.failures.clear();
    await tester.tap(find.text('Reintentar'));
    await tester.pumpAndSettle();

    expect(find.text('PTU pagada'), findsOneWidget);
  });
}
