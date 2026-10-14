import 'dart:async';

import 'package:control_contable/api/generated/models/deductible_source.dart';
import 'package:control_contable/core/access/read_only_access.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/features/expenses/expenses_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../taxpayers/pump_onboarding.dart' show inputLabeled, tapText, typeInto;
import 'expenses_fakes.dart';
import 'pump_expenses.dart';

const _amount = 'Monto sin IVA';
const _description = 'Descripción (opcional)';
const _paidOn = 'Fecha de pago';
const _softId = 'e-soft';

String _textOf(WidgetTester tester, String label) =>
    tester.widget<TextField>(inputLabeled(label)).controller!.text;

void main() {
  late ExpensesHarness harness;

  setUp(() => harness = ExpensesHarness());

  /// Opens the list holding one software expense, then taps its row.
  Future<void> openEdit(
    WidgetTester tester, {
    bool deductible = true,
    DeductibleSource source = DeductibleSource.categoryDefault,
  }) async {
    harness.expenses.expenses.add(
      expense(id: _softId, deductible: deductible, source: source),
    );
    await harness.pump(tester, taxpayer: anaOneRegime);
    await tester.tap(find.text('Suscripción de diseño'));
    await tester.pumpAndSettle();
  }

  group('M4-10 edit', () {
    testWidgets('the same page, prefilled, with one Guardar', (tester) async {
      await openEdit(tester);

      expect(find.text('Editar gasto'), findsOneWidget);
      expect(_textOf(tester, _paidOn), '05/10/2026');
      expect(_textOf(tester, _amount), '500.00');
      expect(find.text('Software y suscripciones'), findsOneWidget);
      expect(find.text(expenseLegend), findsOneWidget);
      expect(find.text('Guardar y crear otro'), findsNothing);
      expect(harness.expenses.categoryRequests.last.on, '2026-10-05');
    });

    testWidgets('sends only what changed (IVA follows the amount)', (
      tester,
    ) async {
      await openEdit(tester);

      await typeInto(tester, _amount, '600');
      await tapText(tester, 'Guardar');

      expect(harness.expenses.patches.single, {
        'amount_cents': 60000,
        'iva_cents': 9600,
      });
      expect(harness.location, ExpensesRoutes.list);
      expect(find.text('Gasto actualizado'), findsOneWidget);
    });

    testWidgets('an edit with no changes sends nothing', (tester) async {
      await openEdit(tester);

      await tapText(tester, 'Guardar');

      expect(harness.expenses.patches, isEmpty);
      expect(harness.location, ExpensesRoutes.list);
    });

    testWidgets('a cleared description is sent as null', (tester) async {
      await openEdit(tester);
      await tapText(tester, 'Más detalles');

      await typeInto(tester, _description, '');
      await tapText(tester, 'Guardar');

      expect(harness.expenses.patches.single, {'description': null});
    });

    testWidgets('a stored override shows as the user\'s; "Usar el valor '
        'sugerido" sends deductible: null', (tester) async {
      await openEdit(tester, deductible: false, source: DeductibleSource.user);
      await tapText(tester, 'Más detalles');
      final toggle = tester.widget<SwitchListTile>(find.byType(SwitchListTile));
      expect(toggle.value, isFalse);

      await tapText(tester, 'Usar el valor sugerido');
      await tapText(tester, 'Guardar');

      expect(harness.expenses.patches.single, {'deductible': null});
    });

    testWidgets('a read-only account gets the paywall on save', (tester) async {
      harness.expenses.failNext['update'] = const ApiEntitlementRequired();
      await openEdit(tester);

      await typeInto(tester, _amount, '600');
      await tapText(tester, 'Guardar');

      expect(find.text('Tu cuenta es de solo consulta'), findsOneWidget);
      expect(harness.container.read(readOnlyAccessProvider), isTrue);
    });

    testWidgets('an expense deleted elsewhere says so', (tester) async {
      harness.expenses.failNext['update'] = const ApiNotFound();
      await openEdit(tester);

      await typeInto(tester, _amount, '600');
      await tapText(tester, 'Guardar');

      expect(
        find.text('Este gasto ya no existe: se eliminó en otro lugar.'),
        findsOneWidget,
      );
    });

    testWidgets('a deep link loads the expense first', (tester) async {
      harness.expenses.expenses.add(expense(id: _softId));
      await harness.pump(tester, taxpayer: anaOneRegime);

      unawaited(
        harness.container
            .read(appRouterProvider)
            .push<void>(ExpensesRoutes.editOf(_softId)),
      );
      await tester.pumpAndSettle();

      expect(harness.expenses.calls, contains('get:$_softId'));
      expect(_textOf(tester, _amount), '500.00');
    });
  });
}
