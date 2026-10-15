import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/annual/annual_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../tax_engine/estimate_fixtures.dart' show fixtureTaxpayerId;
import 'annual_fixtures.dart';
import 'annual_inputs_fixtures.dart';
import 'pump_annual_inputs.dart';

final String _list = AnnualRoutes.personalDeductions(fixtureTaxpayerId, 2026);

/// ANN-19: the personal deductions of a year, grouped by category with
/// registered vs applied, delete + Undo, the 500 bound, read-only.
void main() {
  late AnnualInputsHarness harness;

  setUp(() => harness = AnnualInputsHarness());

  Finder group(String category) =>
      find.byKey(ValueKey('deduction-group-$category'));

  Finder inGroup(String category, Finder finder) =>
      find.descendant(of: group(category), matching: finder);

  testWidgets('groups by category with registered and cash-excluded amounts, '
      'and what the estimate applied with its caps', (tester) async {
    harness.inputs.deductions.addAll([
      deductionJson(),
      deductionJson(
        id: 'ded-2',
        amountCents: 50000,
        paymentMethod: 'cash',
        excluded: 'cash_not_deductible',
        beneficiary: 'descendant',
        paidOn: '2026-05-02',
      ),
      deductionJson(
        id: 'ded-3',
        category: 'donation',
        amountCents: 300000,
        doneeType: 'authorized',
      ),
    ]);
    harness.annual.breakdownFallback = annualBreakdownFrom(
      annualBreakdownJson(
        lines: [
          annualLineJson(1, 'Ingresos acumulables', amountCents: 50000000),
          annualLineJson(
            2,
            r'Tope aplicado: deducciones personales, de $5,000.00 a '
            r'$4,000.00',
            operation: 'info',
          ),
        ],
      ),
    );
    await harness.pump(tester, location: _list);

    expect(find.text('Deducciones personales'), findsWidgets);
    final summary = find.byKey(const ValueKey('deductions-summary'));
    expect(
      find.descendant(of: summary, matching: find.text(r'$5,000.00')),
      findsOneWidget,
    );
    expect(
      find.descendant(of: summary, matching: find.text(r'$40,000.00')),
      findsOneWidget,
    );
    expect(find.text('Tope aplicado'), findsOneWidget);
    expect(find.textContaining(r'de $5,000.00 a $4,000.00'), findsOneWidget);

    expect(
      inGroup('medical', find.text('Honorarios médicos y hospitalarios')),
      findsOneWidget,
    );
    expect(
      inGroup('medical', find.text(r'Registrado: $2,000.00')),
      findsOneWidget,
    );
    expect(
      inGroup(
        'medical',
        find.text(r'No deducible por pago en efectivo: $500.00'),
      ),
      findsOneWidget,
    );
    expect(
      inGroup('medical', find.textContaining('En efectivo: no deducible')),
      findsOneWidget,
    );
    expect(inGroup('medical', find.textContaining('Descendiente')), findsOne);
    expect(inGroup('donation', find.text('Donativos')), findsOneWidget);
    expect(harness.inputs.calls, contains('listDeductions'));
  });

  testWidgets('empty year: says so and offers to add', (tester) async {
    await harness.pump(tester, location: _list);

    expect(
      find.text('Aún no registras deducciones personales de 2026.'),
      findsOneWidget,
    );
    expect(find.text('Agregar deducción'), findsOneWidget);
  });

  testWidgets('a failed load offers a retry', (tester) async {
    harness.inputs.failures['listDeductions'] = const ApiUnreachable();
    await harness.pump(tester, location: _list);

    expect(
      find.text('No se pudieron cargar tus deducciones personales.'),
      findsOneWidget,
    );
    harness.inputs.failures.clear();
    harness.inputs.deductions.add(deductionJson());
    await tester.tap(find.text('Reintentar'));
    await tester.pumpAndSettle();

    expect(group('medical'), findsOneWidget);
  });

  testWidgets('delete hides the row and the Undo toast restores it', (
    tester,
  ) async {
    harness.inputs.deductions.add(deductionJson());
    await harness.pump(tester, location: _list);

    await tester.tap(find.byTooltip(r'Eliminar deducción de $1,500.00'));
    await tester.pumpAndSettle();

    expect(group('medical'), findsNothing);
    expect(find.text('Deducción eliminada'), findsOneWidget);
    expect(harness.inputs.calls, contains('deleteDeduction ded-1'));

    await tester.tap(find.text('Deshacer'));
    await tester.pumpAndSettle();

    expect(harness.inputs.calls, contains('restoreDeduction ded-1'));
    expect(group('medical'), findsOneWidget);
    expect(find.text('Deducción restaurada'), findsOneWidget);
  });

  testWidgets('a refused delete keeps the row and says why', (tester) async {
    harness.inputs.deductions.add(deductionJson());
    harness.inputs.failures['deleteDeduction'] = const ApiUnreachable();
    await harness.pump(tester, location: _list);

    await tester.tap(find.byTooltip(r'Eliminar deducción de $1,500.00'));
    await tester.pumpAndSettle();

    expect(group('medical'), findsOneWidget);
    expect(
      find.text('No se pudo eliminar la deducción. Intenta de nuevo.'),
      findsOneWidget,
    );
  });

  testWidgets('at 500 deductions the year is full: says so, and the add '
      'button explains instead of opening the form', (tester) async {
    harness.inputs.deductions.addAll([
      for (var i = 0; i < 500; i++) deductionJson(id: 'd$i', amountCents: 100),
    ]);
    await harness.pump(tester, location: _list);

    const full =
        'Llegaste al límite de 500 deducciones de 2026. Para agregar otra, '
        'elimina alguna.';
    expect(find.text(full), findsOneWidget);
    await tester.tap(find.text('Agregar deducción'));
    await tester.pumpAndSettle();

    expect(harness.location, _list);
  });

  testWidgets('add opens the form on its own screen', (tester) async {
    await harness.pump(tester, location: _list);

    await tester.tap(find.text('Agregar deducción'));
    await tester.pumpAndSettle();

    expect(
      harness.location,
      AnnualRoutes.newPersonalDeduction(fixtureTaxpayerId, 2026),
    );
  });

  testWidgets('read-only: the list shows, add opens the paywall', (
    tester,
  ) async {
    harness
      ..readOnly = true
      ..inputs.deductions.add(deductionJson());
    await harness.pump(tester, location: _list);

    expect(group('medical'), findsOneWidget);
    await tester.tap(find.text('Agregar deducción'));
    await tester.pumpAndSettle();

    expect(harness.location, _list);
    expect(find.text('Tu cuenta es de solo consulta'), findsOneWidget);
  });
}
