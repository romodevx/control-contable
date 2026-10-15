import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/annual/annual_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../tax_engine/estimate_fixtures.dart' show fixtureTaxpayerId;
import '../taxpayers/taxpayers_fakes.dart' show detail;
import 'annual_inputs_fixtures.dart';
import 'pump_annual_inputs.dart';

final String _list = AnnualRoutes.salaryStatements(fixtureTaxpayerId, 2026);

/// ANN-20: the salary constancias of a year, a sheet form (≤ 6 fields), no
/// employer RFC, delete with confirmation (it cannot be undone).
void main() {
  late AnnualInputsHarness harness;

  setUp(() => harness = AnnualInputsHarness());

  Finder sheet() => find.byKey(const ValueKey('salary-statement-sheet'));

  Future<void> tapText(WidgetTester tester, String text) async {
    final target = find.text(text).last;
    await tester.ensureVisible(target);
    await tester.pumpAndSettle();
    await tester.tap(target);
    await tester.pumpAndSettle();
  }

  /// Types into the sheet's field labeled [label] (the list behind it may
  /// repeat the words).
  Future<void> typeInto(WidgetTester tester, String label, String text) async {
    final field = find.descendant(
      of: find
          .ancestor(
            of: find.descendant(of: sheet(), matching: find.text(label)),
            matching: find.byType(Column),
          )
          .first,
      matching: find.byType(TextField),
    );
    await tester.ensureVisible(field);
    await tester.enterText(field, text);
    await tester.pumpAndSettle();
  }

  Future<void> save(WidgetTester tester) async {
    await tester.tap(find.widgetWithText(FilledButton, 'Guardar'));
    await tester.pumpAndSettle();
  }

  Future<void> fillFigures(WidgetTester tester) async {
    await typeInto(tester, 'Ingresos gravados (acumulables)', '450,000');
    await typeInto(tester, 'Ingresos exentos', '15000');
    await typeInto(tester, 'ISR retenido', '80,000.5');
  }

  testWidgets('lists each constancia with the year totals', (tester) async {
    harness.inputs.statements.addAll([
      statementJson(),
      statementJson(id: 'sal-2', employerLabel: null, isrWithheldCents: 0),
    ]);
    await harness.pump(tester, location: _list);

    expect(find.text('Constancias de sueldos'), findsWidgets);
    expect(find.text('Empresa A'), findsOneWidget);
    expect(find.text('Constancia 2'), findsOneWidget);
    final totals = find.byKey(const ValueKey('salary-totals'));
    expect(
      find.descendant(of: totals, matching: find.text(r'$900,000.00')),
      findsOneWidget,
    );
    expect(
      find.descendant(of: totals, matching: find.text(r'$80,000.00')),
      findsOneWidget,
    );
  });

  testWidgets('empty year and a failed load', (tester) async {
    harness.inputs.failures['listStatements'] = const ApiUnreachable();
    await harness.pump(tester, location: _list);

    expect(
      find.text('No se pudieron cargar tus constancias de sueldos.'),
      findsOneWidget,
    );
    harness.inputs.failures.clear();
    await tester.tap(find.text('Reintentar'));
    await tester.pumpAndSettle();

    expect(
      find.text('Aún no registras constancias de sueldos de 2026.'),
      findsOneWidget,
    );
  });

  testWidgets('add: a sheet with the constancia totals and no employer RFC; '
      'saves and closes', (tester) async {
    await harness.pump(tester, location: _list);
    await tapText(tester, 'Agregar constancia');

    expect(sheet(), findsOneWidget);
    expect(find.text('Nueva constancia'), findsOneWidget);
    expect(find.textContaining('RFC del empleador'), findsOneWidget);
    expect(find.widgetWithText(TextField, 'RFC'), findsNothing);
    await typeInto(tester, 'Nombre del empleador (opcional)', 'Empresa A');
    await fillFigures(tester);
    await save(tester);

    final sent = harness.inputs.createdStatements.single;
    expect(sent.id, 'client-1');
    expect(sent.fiscalYear, 2026);
    expect(sent.employerLabel, 'Empresa A');
    expect(sent.accumulableCents, 45000000);
    expect(sent.exemptCents, 1500000);
    expect(sent.isrWithheldCents, 8000050);
    expect(sent.simultaneous, isFalse);
    expect(sheet(), findsNothing);
    expect(find.text('Constancia guardada'), findsOneWidget);
    expect(find.text('Empresa A'), findsOneWidget);
  });

  testWidgets('required figures are asked on submit only', (tester) async {
    await harness.pump(tester, location: _list);
    await tapText(tester, 'Agregar constancia');
    await save(tester);

    expect(harness.inputs.createdStatements, isEmpty);
    expect(find.text('Escribe este dato.'), findsWidgets);
    expect(sheet(), findsOneWidget);
  });

  testWidgets('special situations sit under a disclosure', (tester) async {
    await harness.pump(tester, location: _list);
    await tapText(tester, 'Agregar constancia');
    await fillFigures(tester);

    expect(
      find.text('Trabajé para dos o más empleadores al mismo tiempo'),
      findsNothing,
    );
    await tapText(tester, 'Situaciones especiales');
    await tapText(tester, 'Trabajé para dos o más empleadores al mismo tiempo');
    await save(tester);

    expect(harness.inputs.createdStatements.single.simultaneous, isTrue);
  });

  testWidgets('a constancia without regime 605 saves with its suggestion', (
    tester,
  ) async {
    harness.inputs.statementWarnings.add({
      'code': 'salary_regime_not_held',
      'field': 'fiscal_year',
      'message': 'Add 605.',
    });
    await harness.pump(tester, location: _list);
    await tapText(tester, 'Agregar constancia');
    await fillFigures(tester);
    await save(tester);

    expect(find.textContaining('régimen de sueldos (605)'), findsOneWidget);
  });

  testWidgets('edit: the row opens the sheet prefilled; only changes are '
      'sent, a cleared name is null', (tester) async {
    harness.inputs.statements.add(statementJson());
    await harness.pump(tester, location: _list);
    await tapText(tester, 'Empresa A');

    expect(find.text('Editar constancia'), findsOneWidget);
    await typeInto(tester, 'Nombre del empleador (opcional)', '');
    await typeInto(tester, 'ISR retenido', '81,000');
    await save(tester);

    expect(harness.inputs.statementPatches.single, {
      'employer_label': null,
      'isr_withheld_cents': 8100000,
    });
    expect(find.text('Constancia actualizada'), findsOneWidget);
  });

  testWidgets('delete asks first; keeping it does nothing', (tester) async {
    harness.inputs.statements.add(statementJson());
    await harness.pump(tester, location: _list);

    await tester.tap(find.byTooltip('Eliminar Empresa A'));
    await tester.pumpAndSettle();
    expect(find.text('¿Eliminar Empresa A?'), findsOneWidget);
    await tester.tap(find.text('Conservar'));
    await tester.pumpAndSettle();
    expect(harness.inputs.calls, isNot(contains('deleteStatement sal-1')));

    await tester.tap(find.byTooltip('Eliminar Empresa A'));
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(FilledButton, 'Eliminar'));
    await tester.pumpAndSettle();

    expect(harness.inputs.calls, contains('deleteStatement sal-1'));
    expect(find.text('Constancia eliminada'), findsOneWidget);
    expect(find.text('Empresa A'), findsNothing);
  });

  testWidgets('the 20 bound refuses with its message', (tester) async {
    harness.inputs.failures['createStatement'] = ApiRejected(
      code: ErrorCode.businessRuleViolation,
      details: [detail('body', 'limit_reached')],
    );
    await harness.pump(tester, location: _list);
    await tapText(tester, 'Agregar constancia');
    await fillFigures(tester);
    await save(tester);

    expect(
      find.text('Llegaste al límite de 20 constancias de este año.'),
      findsOneWidget,
    );
    expect(sheet(), findsOneWidget);
  });

  testWidgets('read-only: add opens the paywall; a 402 on save too', (
    tester,
  ) async {
    harness.readOnly = true;
    await harness.pump(tester, location: _list);
    await tapText(tester, 'Agregar constancia');

    expect(sheet(), findsNothing);
    expect(find.text('Tu cuenta es de solo consulta'), findsOneWidget);
  });

  testWidgets('a 402 on save opens the paywall over the sheet', (tester) async {
    harness.inputs.failures['createStatement'] = const ApiEntitlementRequired();
    await harness.pump(tester, location: _list);
    await tapText(tester, 'Agregar constancia');
    await fillFigures(tester);
    await save(tester);

    expect(find.text('Tu cuenta es de solo consulta'), findsOneWidget);
  });
}
