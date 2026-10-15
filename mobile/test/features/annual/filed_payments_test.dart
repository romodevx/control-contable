import 'package:control_contable/api/generated/models/filed_payment_entry.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/annual/annual_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../tax_engine/estimate_fixtures.dart' show fixtureTaxpayerId;
import '../taxpayers/pump_onboarding.dart' show inputLabeled, typeInto;
import 'annual_inputs_fixtures.dart';
import 'pump_annual_inputs.dart';

final String _screen = AnnualRoutes.filedPayments(
  fixtureTaxpayerId,
  2026,
  '612',
);

/// January confirmed, February from the tax input, March and April
/// estimated (April from an incomplete estimate), May with nothing to
/// prefill, June–November estimated, December outside the regime.
List<Map<String, Object?>> _months() => [
  paymentMonthJson(1, filedCents: 120000),
  paymentMonthJson(2, taxInputCents: 90000),
  paymentMonthJson(3),
  paymentMonthJson(4, appEstimateIncomplete: true),
  paymentMonthJson(5, appEstimateCents: null),
  for (var m = 6; m <= 11; m++) paymentMonthJson(m),
  paymentMonthJson(12, inWindow: false, appEstimateCents: null),
];

/// ANN-21: the twelve provisional payments of one regime as inline rows,
/// prefilled with what the app knows, and "Confirmar todos".
void main() {
  late AnnualInputsHarness harness;

  setUp(() {
    harness = AnnualInputsHarness();
    harness.inputs.payments['612'] = filedPaymentsOf(months: _months());
  });

  String textOf(WidgetTester tester, String month) =>
      tester.widget<TextField>(inputLabeled(month)).controller!.text;

  Map<int, int> sent() => {
    for (final FiledPaymentEntry e
        in harness.inputs.putPaymentBodies.single.$2.months)
      e.month: e.amountCents,
  };

  Future<void> tapButton(WidgetTester tester, String label) async {
    final button = find.text(label).first;
    await tester.ensureVisible(button);
    await tester.pumpAndSettle();
    await tester.tap(button);
    await tester.pumpAndSettle();
  }

  testWidgets('one row per month, prefilled and labelled by its source', (
    tester,
  ) async {
    await harness.pump(tester, location: _screen);

    expect(find.text('Pagos provisionales presentados'), findsOneWidget);
    expect(find.text('Régimen 612'), findsOneWidget);
    expect(textOf(tester, 'Enero'), '1,200.00');
    expect(find.text('Confirmado'), findsOneWidget);
    expect(textOf(tester, 'Febrero'), '900.00');
    expect(find.text('De tus datos fiscales — confírmalo'), findsOneWidget);
    expect(textOf(tester, 'Marzo'), '1,000.00');
    expect(find.text('Estimado por la app — confírmalo'), findsNWidgets(7));
    expect(
      find.text('Estimado con datos incompletos — confírmalo'),
      findsOneWidget,
    );
    expect(textOf(tester, 'Mayo'), '');
    expect(find.text('Sin estimación: escribe lo que pagaste'), findsOneWidget);
    await tester.ensureVisible(find.text('Diciembre'));
    await tester.pumpAndSettle();
    expect(find.text('No tenías este régimen en este mes.'), findsOneWidget);
    expect(inputLabeled('Diciembre'), findsNothing);
  });

  testWidgets('"Confirmar todos" stores every shown amount as filed', (
    tester,
  ) async {
    await harness.pump(tester, location: _screen);
    await tapButton(tester, 'Confirmar todos');

    expect(sent(), {
      1: 120000,
      2: 90000,
      3: 100000,
      4: 100000,
      for (var m = 6; m <= 11; m++) m: 100000,
    });
    expect(find.text('Pagos guardados: 10 meses confirmados'), findsOneWidget);
    expect(find.text('Confirmado'), findsNWidgets(10));
  });

  testWidgets('Guardar keeps confirmed months and adds only the edited ones', (
    tester,
  ) async {
    await harness.pump(tester, location: _screen);
    await typeInto(tester, 'Marzo', '1,050.5');
    await typeInto(tester, 'Mayo', '700');
    await tapButton(tester, 'Guardar');

    expect(sent(), {1: 120000, 3: 105050, 5: 70000});
  });

  testWidgets('clearing a confirmed month un-confirms it', (tester) async {
    await harness.pump(tester, location: _screen);
    await typeInto(tester, 'Enero', '');
    await tapButton(tester, 'Guardar');

    expect(sent(), isEmpty);
  });

  testWidgets('a malformed amount stops the save', (tester) async {
    await harness.pump(tester, location: _screen);
    await typeInto(tester, 'Marzo', '1,23');
    await tapButton(tester, 'Guardar');

    expect(harness.inputs.putPaymentBodies, isEmpty);
    expect(find.textContaining('Revisa'), findsWidgets);
  });

  testWidgets('read-only: Confirmar todos opens the paywall, nothing sent', (
    tester,
  ) async {
    harness.readOnly = true;
    await harness.pump(tester, location: _screen);
    await tapButton(tester, 'Confirmar todos');

    expect(harness.inputs.putPaymentBodies, isEmpty);
    expect(find.text('Tu cuenta es de solo consulta'), findsOneWidget);
  });

  testWidgets('a 402 opens the paywall and keeps what was typed', (
    tester,
  ) async {
    harness.inputs.failures['putFiledPayments'] =
        const ApiEntitlementRequired();
    await harness.pump(tester, location: _screen);
    await typeInto(tester, 'Marzo', '1,050');
    await tapButton(tester, 'Guardar');

    expect(find.text('Tu cuenta es de solo consulta'), findsOneWidget);
    await tester.tapAt(Offset.zero);
    await tester.pumpAndSettle();
    expect(textOf(tester, 'Marzo'), '1,050.00');
  });

  testWidgets('a failed load offers a retry', (tester) async {
    harness.inputs.failures['getFiledPayments'] = const ApiUnreachable();
    await harness.pump(tester, location: _screen);

    expect(
      find.text('No se pudieron cargar tus pagos provisionales.'),
      findsOneWidget,
    );
    harness.inputs.failures.clear();
    await tester.tap(find.text('Reintentar'));
    await tester.pumpAndSettle();

    expect(inputLabeled('Enero'), findsOneWidget);
  });
}
