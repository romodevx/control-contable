import 'dart:async';

import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/core/access/read_only_access.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/features/expenses/expenses_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../taxpayers/taxpayers_fakes.dart';
import 'expenses_fakes.dart';
import 'pump_expenses.dart';

const _paywallTitle = 'Tu cuenta es de solo consulta';

Finder _deleteButtonOf(String title) =>
    find.byTooltip('Eliminar gasto: $title');

/// Taps the delete button of [title] on a phone-tall surface (on the
/// default 800×600 one the floating button covers the last row).
Future<void> _delete(WidgetTester tester, String title) async {
  tester.view.physicalSize = const Size(2400, 4200);
  addTearDown(tester.view.resetPhysicalSize);
  await tester.pumpAndSettle();
  await tester.tap(_deleteButtonOf(title));
  await tester.pumpAndSettle();
}

void main() {
  late ExpensesHarness harness;

  setUp(() => harness = ExpensesHarness());

  void seed() => harness.expenses.expenses.addAll([
    expense(id: 'e-soft'),
    expense(
      id: 'e-rest',
      regime: '606',
      category: 'restaurantes',
      description: null,
      amountCents: 100000,
      ivaCents: 16000,
      paidOn: '2026-10-09',
      deductible: false,
    ),
    expense(
      id: 'e-sep',
      description: 'Papel',
      category: 'papeleria',
      paidOn: '2026-09-20',
    ),
  ]);

  group('M4-09 list', () {
    testWidgets('opens on the month the location names (History, US-H6)', (
      tester,
    ) async {
      seed();
      await harness.pump(tester, location: ExpensesRoutes.listOf('2026-09'));

      expect(find.text('septiembre de 2026'), findsOneWidget);
      expect(harness.expenses.calls.first, 'list:2026-09 * -');
      expect(find.text('Papel'), findsOneWidget);
    });

    testWidgets('a malformed or future month in the location opens the '
        'current one', (tester) async {
      seed();
      await harness.pump(tester, location: '/gastos?periodo=2026-11');

      expect(find.text('octubre de 2026'), findsOneWidget);
      expect(harness.expenses.calls.first, 'list:2026-10 * -');
    });

    testWidgets('shows the current month newest first: date, what it was, '
        'total with IVA, regime, a "No deducible" badge, and the month '
        'totals', (tester) async {
      seed();
      await harness.pump(tester);

      expect(find.text('octubre de 2026'), findsOneWidget);
      expect(harness.expenses.calls, contains('list:2026-10 * -'));
      expect(find.text('Suscripción de diseño'), findsOneWidget);
      // No description: the category's name.
      expect(find.text('Consumo en restaurantes'), findsOneWidget);
      expect(find.text(r'$580.00'), findsOneWidget);
      expect(find.text(r'$1,160.00'), findsOneWidget);
      expect(find.text('9 oct'), findsOneWidget);
      expect(find.text('Régimen 606'), findsWidgets);
      expect(find.text('No deducible'), findsOneWidget);
      // 500 + 1,000 + IVA 80 + 160.
      expect(find.text(r'$1,740.00'), findsOneWidget);
      expect(find.text('2 gastos'), findsOneWidget);
      final restaurant = tester.getTopLeft(
        find.text('Consumo en restaurantes'),
      );
      final software = tester.getTopLeft(find.text('Suscripción de diseño'));
      expect(restaurant.dy, lessThan(software.dy));
    });

    testWidgets('the mandatory legend heads the list, from the server field', (
      tester,
    ) async {
      seed();
      await harness.pump(tester);

      final legend = find.text(expenseLegend);
      expect(legend, findsOneWidget);
      expect(
        tester.getTopLeft(legend).dy,
        lessThan(tester.getTopLeft(find.text('2 gastos')).dy),
      );
    });

    testWidgets('an empty month still shows the legend and offers to record '
        'one', (tester) async {
      await harness.pump(tester);

      expect(find.text('Sin gastos en octubre de 2026'), findsOneWidget);
      expect(find.text(expenseLegend), findsOneWidget);
      expect(find.text('Registrar gasto'), findsWidgets);
    });

    testWidgets('moves between months and filters by regime', (tester) async {
      seed();
      await harness.pump(tester);

      await tester.tap(find.byTooltip('Mes anterior'));
      await tester.pumpAndSettle();
      expect(find.text('Papel'), findsOneWidget);
      expect(harness.expenses.calls, contains('list:2026-09 * -'));

      await tester.tap(find.byTooltip('Mes siguiente'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(ChoiceChip, 'Régimen 612'));
      await tester.pumpAndSettle();

      expect(harness.expenses.calls, contains('list:2026-10 612 -'));
      expect(find.text('Suscripción de diseño'), findsOneWidget);
      expect(find.text('Consumo en restaurantes'), findsNothing);
      expect(find.text('1 gasto'), findsOneWidget);
    });

    testWidgets('a failed load offers a retry', (tester) async {
      harness.expenses.failNext['list'] = const ApiUnreachable();
      await harness.pump(tester);

      expect(find.text('No se pudieron cargar los gastos.'), findsOneWidget);
      await tester.tap(find.text('Reintentar'));
      await tester.pumpAndSettle();
      expect(find.text('Sin gastos en octubre de 2026'), findsOneWidget);
    });

    testWidgets('tapping a row opens its edit page', (tester) async {
      seed();
      await harness.pump(tester);

      await tester.tap(find.text('Suscripción de diseño'));
      await tester.pumpAndSettle();

      expect(harness.location, ExpensesRoutes.editOf('e-soft'));
      expect(find.text('Editar gasto'), findsOneWidget);
    });
  });

  group('M4-09 delete and Undo', () {
    testWidgets('delete hides the row at once; the Undo toast never '
        'auto-dismisses and restores the expense', (tester) async {
      seed();
      await harness.pump(tester);

      await _delete(tester, 'Suscripción de diseño');

      expect(harness.expenses.calls, contains('delete:e-soft'));
      expect(find.text('Suscripción de diseño'), findsNothing);
      expect(find.text('1 gasto'), findsOneWidget);
      expect(find.text('Gasto eliminado'), findsOneWidget);

      await tester.pump(const Duration(minutes: 5));
      await tester.pumpAndSettle();
      final snackBar = tester.widget<SnackBar>(find.byType(SnackBar));
      expect(snackBar.persist, isTrue);
      expect(snackBar.showCloseIcon, isTrue);

      await tester.tap(find.text('Deshacer'));
      await tester.pumpAndSettle();

      expect(harness.expenses.calls, contains('restore:e-soft'));
      expect(find.text('Suscripción de diseño'), findsOneWidget);
      expect(find.text('Gasto restaurado'), findsOneWidget);
    });

    testWidgets('a read-only user gets the same Undo, and restoring never '
        'opens the paywall (contract 0.6.1)', (tester) async {
      seed();
      await harness.pump(tester);
      harness.container.read(readOnlyAccessProvider.notifier).markReadOnly();

      await _delete(tester, 'Suscripción de diseño');
      expect(find.byType(AlertDialog), findsNothing);
      await tester.tap(find.text('Deshacer'));
      await tester.pumpAndSettle();

      expect(harness.expenses.calls, contains('restore:e-soft'));
      expect(find.text(_paywallTitle), findsNothing);
    });

    testWidgets('a failed delete while an Undo is pending keeps the Undo '
        'up (PR #85 follow-up)', (tester) async {
      seed();
      await harness.pump(tester);

      await _delete(tester, 'Suscripción de diseño');
      harness.expenses.failNext['delete'] = const ApiUnreachable();
      await _delete(tester, 'Consumo en restaurantes');

      expect(find.text('Consumo en restaurantes'), findsOneWidget);
      expect(find.text('Gasto eliminado'), findsOneWidget);
      await tester.tap(find.text('Deshacer'));
      await tester.pumpAndSettle();
      expect(harness.expenses.calls, contains('restore:e-soft'));
      expect(find.text('No se pudo eliminar el gasto.'), findsOneWidget);
    });

    testWidgets('a stacked Undo restores every expense it can and says how '
        'many came back (PR #85 follow-up)', (tester) async {
      seed();
      harness.expenses.failRestoreOf['e-soft'] = ApiRejected(
        code: ErrorCode.businessRuleViolation,
        details: [detail('regime_code', 'regime_not_active')],
      );
      await harness.pump(tester);

      await _delete(tester, 'Suscripción de diseño');
      await _delete(tester, 'Consumo en restaurantes');
      expect(find.text('2 gastos eliminados'), findsOneWidget);

      await tester.tap(find.text('Deshacer'));
      await tester.pumpAndSettle();

      expect(harness.expenses.calls, contains('restore:e-rest'));
      expect(find.text('Consumo en restaurantes'), findsOneWidget);
      expect(
        find.text(
          'Se restauró 1 de 2 gastos. No se pudo restaurar: el régimen ya no '
          'cubre la fecha del gasto.',
        ),
        findsOneWidget,
      );
    });
  });

  group('M4-09 flag', () {
    testWidgets('module_expenses off: no entry on home, the list is not '
        'found and nothing loads', (tester) async {
      await harness.pump(tester, enabled: false, location: null);

      expect(find.text('Gastos'), findsNothing);
      unawaited(
        harness.container
            .read(appRouterProvider)
            .push<void>(ExpensesRoutes.list),
      );
      await tester.pumpAndSettle();
      expect(harness.location, AppRoutes.notFound);
      expect(harness.expenses.calls, isEmpty);
    });

    testWidgets('module_expenses on: home leads to the list', (tester) async {
      await harness.pump(tester, location: null);

      await tester.tap(find.widgetWithText(OutlinedButton, 'Gastos'));
      await tester.pumpAndSettle();
      expect(harness.location, ExpensesRoutes.list);
    });
  });

  for (final brightness in Brightness.values) {
    testWidgets('meets the tap-target, label and contrast guidelines '
        '(${brightness.name})', (tester) async {
      final handle = tester.ensureSemantics();
      seed();
      await harness.pump(tester, brightness: brightness);

      await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
      await expectLater(tester, meetsGuideline(iOSTapTargetGuideline));
      await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
      await expectLater(tester, meetsGuideline(textContrastGuideline));
      handle.dispose();
    });
  }
}
