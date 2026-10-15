import 'dart:async';

import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/core/access/read_only_access.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/features/incomes/incomes_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../taxpayers/taxpayers_fakes.dart';
import 'incomes_fakes.dart';
import 'pump_incomes.dart';

const _paywallTitle = 'Tu cuenta es de solo consulta';

Finder _deleteButtonOf(String description) =>
    find.byTooltip('Eliminar ingreso: $description');

void main() {
  late IncomesHarness harness;

  setUp(() => harness = IncomesHarness());

  void seed() => harness.incomes.incomes.addAll([
    income(id: 'i-logo'),
    income(
      id: 'i-rent',
      regime: '606',
      description: 'Renta de octubre',
      amountCents: 1500000,
      ivaCents: 0,
      collectedOn: '2026-10-09',
    ),
    income(
      id: 'i-sep',
      description: 'Asesoría',
      amountCents: 500000,
      ivaCents: 80000,
      collectedOn: '2026-09-20',
    ),
  ]);

  group('M3-08 list', () {
    testWidgets('opens on the month the location names (History, US-H6)', (
      tester,
    ) async {
      seed();
      await harness.pump(tester, location: IncomesRoutes.listOf('2026-09'));

      expect(find.text('septiembre de 2026'), findsOneWidget);
      expect(harness.incomes.calls.first, 'list:2026-09 * -');
      expect(find.text('Asesoría'), findsOneWidget);
    });

    testWidgets('a malformed or future month in the location opens the '
        'current one', (tester) async {
      seed();
      await harness.pump(tester, location: '/ingresos?periodo=2026-11');

      expect(find.text('octubre de 2026'), findsOneWidget);
      expect(harness.incomes.calls.first, 'list:2026-10 * -');
    });

    testWidgets('shows the current month newest first: date, description, '
        'total with IVA and regime, plus the month totals in es-MX', (
      tester,
    ) async {
      seed();
      await harness.pump(tester);

      expect(find.text('octubre de 2026'), findsOneWidget);
      expect(harness.incomes.calls.first, 'list:2026-10 * -');
      expect(find.text('Renta de octubre'), findsOneWidget);
      expect(find.text('Diseño de logotipo'), findsOneWidget);
      expect(find.text(r'$11,600.00'), findsOneWidget);
      expect(find.text(r'$15,000.00'), findsWidgets);
      expect(find.text('9 oct'), findsOneWidget);
      expect(find.text('Régimen 606'), findsWidgets);
      // Totals cover the whole month: 10,000 + 15,000 + IVA 1,600.
      expect(find.text(r'$26,600.00'), findsOneWidget);
      expect(find.text('2 ingresos'), findsOneWidget);
      final rent = tester.getTopLeft(find.text('Renta de octubre'));
      final logo = tester.getTopLeft(find.text('Diseño de logotipo'));
      expect(rent.dy, lessThan(logo.dy));
    });

    testWidgets('moves between months; the current month is the last one', (
      tester,
    ) async {
      seed();
      await harness.pump(tester);

      await tester.tap(find.byTooltip('Mes anterior'));
      await tester.pumpAndSettle();

      expect(find.text('septiembre de 2026'), findsOneWidget);
      expect(harness.incomes.calls.last, 'list:2026-09 * -');
      expect(find.text('Asesoría'), findsOneWidget);
      expect(find.text('Renta de octubre'), findsNothing);

      await tester.tap(find.byTooltip('Mes siguiente'));
      await tester.pumpAndSettle();
      expect(find.text('octubre de 2026'), findsOneWidget);
      final next = tester.widget<IconButton>(
        find.ancestor(
          of: find.byTooltip('Mes siguiente'),
          matching: find.byType(IconButton),
        ),
      );
      expect(next.onPressed, isNull);
    });

    testWidgets('an empty month says so warmly and offers to record one', (
      tester,
    ) async {
      await harness.pump(tester);

      expect(find.text('Sin ingresos en octubre de 2026'), findsOneWidget);
      expect(
        find.text(
          'Registra tu primer ingreso y te mostramos cuánto podrías pagar.',
        ),
        findsOneWidget,
      );

      await tester.tap(find.widgetWithText(FilledButton, 'Registrar ingreso'));
      await tester.pumpAndSettle();
      expect(harness.location, IncomesRoutes.create);
    });

    testWidgets('the regime filter narrows the items and the totals', (
      tester,
    ) async {
      seed();
      await harness.pump(tester);

      await tester.tap(find.widgetWithText(ChoiceChip, 'Régimen 606'));
      await tester.pumpAndSettle();

      expect(harness.incomes.calls.last, 'list:2026-10 606 -');
      expect(find.text('Diseño de logotipo'), findsNothing);
      expect(find.text('1 ingreso'), findsOneWidget);

      await tester.tap(find.widgetWithText(ChoiceChip, 'Todos'));
      await tester.pumpAndSettle();
      expect(find.text('Diseño de logotipo'), findsOneWidget);
    });

    testWidgets('a taxpayer with one regime gets no filter', (tester) async {
      seed();
      await harness.pump(tester, taxpayer: anaOneRegime);

      expect(find.byType(ChoiceChip), findsNothing);
    });

    testWidgets('a row without description names who paid; an accrual one '
        'not yet collected reads "Por cobrar"', (tester) async {
      harness.incomes.incomes.add(
        income(
          regime: '601',
          description: null,
          invoicedOn: '2026-10-02',
          collectedOn: null,
        ),
      );
      await harness.pump(
        tester,
        taxpayer: taxpayer(regimes: [heldRegime('601')]),
      );

      expect(find.text('Persona moral'), findsOneWidget);
      expect(find.text('Por cobrar'), findsOneWidget);
    });

    testWidgets('loads the next page when the list reaches its end', (
      tester,
    ) async {
      harness.incomes.pageSize = 2;
      for (var day = 1; day <= 5; day++) {
        harness.incomes.incomes.add(
          income(description: 'Cobro $day', collectedOn: '2026-10-0$day'),
        );
      }
      await harness.pump(tester);

      await tester.scrollUntilVisible(find.text('Cobro 1'), 200);
      await tester.pumpAndSettle();

      expect(find.text('Cobro 1'), findsOneWidget);
      expect(harness.incomes.calls, contains('list:2026-10 * 4'));
      await tester.scrollUntilVisible(find.text('5 ingresos'), -200);
      expect(find.text('5 ingresos'), findsOneWidget);
      expect(harness.incomes.calls.where((c) => c.startsWith('list')), [
        'list:2026-10 * -',
        'list:2026-10 * 2',
        'list:2026-10 * 4',
      ]);
    });

    testWidgets('a failed load offers a retry', (tester) async {
      seed();
      harness.incomes.failNext['list'] = const ApiUnreachable();
      await harness.pump(tester);

      expect(find.text('No se pudieron cargar los ingresos.'), findsOneWidget);
      await tester.tap(find.text('Reintentar'));
      await tester.pumpAndSettle();
      expect(find.text('Diseño de logotipo'), findsOneWidget);
    });

    testWidgets('a taxpayer gone elsewhere (404) reloads the taxpayers', (
      tester,
    ) async {
      harness.incomes.failNext['list'] = const ApiNotFound();
      await harness.pump(tester);

      expect(
        harness.taxpayers.calls.where((c) => c.startsWith('list')),
        hasLength(2),
      );
    });
  });

  group('M3-08 delete and Undo', () {
    testWidgets('delete hides the row at once; the Undo toast stays until '
        'dismissed and restores the income', (tester) async {
      seed();
      await harness.pump(tester);

      await tester.tap(_deleteButtonOf('Diseño de logotipo'));
      await tester.pumpAndSettle();

      expect(harness.incomes.calls, contains('delete:i-logo'));
      expect(find.text('Diseño de logotipo'), findsNothing);
      expect(find.text('1 ingreso'), findsOneWidget);
      expect(find.text('Ingreso eliminado'), findsOneWidget);

      // Never auto-dismissed.
      await tester.pump(const Duration(minutes: 5));
      await tester.pumpAndSettle();
      expect(find.text('Ingreso eliminado'), findsOneWidget);

      await tester.tap(find.text('Deshacer'));
      await tester.pumpAndSettle();

      expect(harness.incomes.calls, contains('restore:i-logo'));
      expect(find.text('Diseño de logotipo'), findsOneWidget);
      expect(find.text('2 ingresos'), findsOneWidget);
    });

    testWidgets('the toast is a polite status region with a close button', (
      tester,
    ) async {
      seed();
      await harness.pump(tester);

      await tester.tap(_deleteButtonOf('Diseño de logotipo'));
      await tester.pumpAndSettle();

      final snackBar = tester.widget<SnackBar>(find.byType(SnackBar));
      expect(snackBar.persist, isTrue);
      expect(snackBar.showCloseIcon, isTrue);

      await tester.tap(find.byIcon(Icons.close));
      await tester.pumpAndSettle();
      expect(find.text('Ingreso eliminado'), findsNothing);
      expect(harness.incomes.calls, isNot(contains('restore:i-logo')));
    });

    testWidgets('several deletes stack into one Undo that restores them all', (
      tester,
    ) async {
      seed();
      await harness.pump(tester);

      await tester.tap(_deleteButtonOf('Diseño de logotipo'));
      await tester.pumpAndSettle();
      await tester.tap(_deleteButtonOf('Renta de octubre'));
      await tester.pumpAndSettle();

      expect(find.text('2 ingresos eliminados'), findsOneWidget);

      await tester.tap(find.text('Deshacer'));
      await tester.pumpAndSettle();

      expect(
        harness.incomes.calls.where((c) => c.startsWith('restore')),
        unorderedEquals(['restore:i-logo', 'restore:i-rent']),
      );
      expect(find.text('Diseño de logotipo'), findsOneWidget);
      expect(find.text('Renta de octubre'), findsOneWidget);
    });

    testWidgets('a failed delete brings the row back and says so', (
      tester,
    ) async {
      seed();
      harness.incomes.failNext['delete'] = const ApiUnreachable();
      await harness.pump(tester);

      await tester.tap(_deleteButtonOf('Diseño de logotipo'));
      await tester.pumpAndSettle();

      expect(find.text('Diseño de logotipo'), findsOneWidget);
      expect(find.text('No se pudo eliminar el ingreso.'), findsOneWidget);
    });

    testWidgets('an Undo whose regime no longer covers the date says why', (
      tester,
    ) async {
      seed();
      harness.incomes.failNext['restore'] = ApiRejected(
        code: ErrorCode.businessRuleViolation,
        details: [detail('regime_code', 'regime_not_active')],
      );
      await harness.pump(tester);

      await tester.tap(_deleteButtonOf('Diseño de logotipo'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Deshacer'));
      await tester.pumpAndSettle();

      expect(
        find.text(
          'No se pudo restaurar: el régimen ya no cubre la fecha del ingreso.',
        ),
        findsOneWidget,
      );
    });

    // Tech-lead gate M3-12: the M4 rule (PR #85 follow-up) was not carried
    // back; one refused income stranded the rest of the stack.
    testWidgets('a stacked Undo restores every income it can and says how '
        'many came back', (tester) async {
      seed();
      harness.incomes.failNext['restore'] = ApiRejected(
        code: ErrorCode.businessRuleViolation,
        details: [detail('regime_code', 'regime_not_active')],
      );
      await harness.pump(tester);

      await tester.tap(_deleteButtonOf('Diseño de logotipo'));
      await tester.pumpAndSettle();
      await tester.tap(_deleteButtonOf('Renta de octubre'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Deshacer'));
      await tester.pumpAndSettle();

      expect(harness.incomes.calls.where((c) => c.startsWith('restore')), [
        'restore:i-logo',
        'restore:i-rent',
      ]);
      expect(find.text('Renta de octubre'), findsOneWidget);
      expect(
        find.text(
          'Se restauró 1 de 2 ingresos. No se pudo restaurar: el régimen ya '
          'no cubre la fecha del ingreso.',
        ),
        findsOneWidget,
      );
    });

    testWidgets('a read-only user deletes with no confirmation and gets the '
        'same persistent Undo, which restores the income (contract 0.6.1)', (
      tester,
    ) async {
      seed();
      await harness.pump(tester);
      harness.container.read(readOnlyAccessProvider.notifier).markReadOnly();

      await tester.tap(_deleteButtonOf('Diseño de logotipo'));
      await tester.pumpAndSettle();

      expect(find.byType(AlertDialog), findsNothing);
      expect(harness.incomes.calls, contains('delete:i-logo'));
      final snackBar = tester.widget<SnackBar>(find.byType(SnackBar));
      expect(snackBar.persist, isTrue);

      await tester.tap(find.text('Deshacer'));
      await tester.pumpAndSettle();

      expect(harness.incomes.calls, contains('restore:i-logo'));
      expect(find.text('Diseño de logotipo'), findsOneWidget);
      expect(find.text(_paywallTitle), findsNothing);
      expect(harness.container.read(readOnlyAccessProvider), isTrue);
    });

    testWidgets('an Undo refused for an unexpected reason says so and '
        'never opens the paywall', (tester) async {
      seed();
      harness.incomes.failNext['restore'] = const ApiEntitlementRequired();
      await harness.pump(tester);

      await tester.tap(_deleteButtonOf('Diseño de logotipo'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Deshacer'));
      await tester.pumpAndSettle();

      expect(find.text(_paywallTitle), findsNothing);
      expect(find.text('No se pudo restaurar el ingreso.'), findsOneWidget);
      expect(harness.container.read(readOnlyAccessProvider), isFalse);
    });
  });

  group('M3-08 flag', () {
    testWidgets('module_incomes off: no entry on home, the list is not '
        'found and nothing loads', (tester) async {
      await harness.pump(tester, enabled: false, location: null);

      expect(find.text('Ingresos'), findsNothing);
      unawaited(
        harness.container
            .read(appRouterProvider)
            .push<void>(IncomesRoutes.list),
      );
      await tester.pumpAndSettle();
      expect(harness.location, AppRoutes.notFound);
      expect(harness.incomes.calls, isEmpty);
    });

    testWidgets('module_incomes on: home leads to the list', (tester) async {
      await harness.pump(tester, location: null);

      await tester.tap(find.widgetWithText(OutlinedButton, 'Ingresos'));
      await tester.pumpAndSettle();
      expect(harness.location, IncomesRoutes.list);
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
