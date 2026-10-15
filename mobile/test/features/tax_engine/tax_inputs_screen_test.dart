import 'dart:async';
import 'dart:ui' show Tristate;

import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/api/generated/models/field_error.dart';
import 'package:control_contable/api/generated/models/tax_input_value.dart';
import 'package:control_contable/api/generated/models/tax_inputs.dart';
import 'package:control_contable/core/access/read_only_access.dart';
import 'package:control_contable/core/fiscal/tax_data_revision.dart';
import 'package:control_contable/core/forms/app_checkbox_field.dart';
import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/tax_engine/tax_engine_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'estimate_fixtures.dart';
import 'pump_estimate.dart';
import 'tax_inputs_fixtures.dart';

const String _priorPayments = 'prior_provisional_payments';

/// A small phone in portrait (logical pixels).
const Size _phone = Size(360, 740);

String _inputs(String regime, {int fiscalYear = 2026, String? field}) =>
    TaxEngineRoutes.taxInputs(
      fixtureTaxpayerId,
      regime,
      fiscalYear: fiscalYear,
      field: field ?? _priorPayments,
    );

Finder _field(String label) => find.byWidgetPredicate(
  (widget) => widget is AppTextField && widget.field.label == label,
);

Finder _monthFields() => find.byWidgetPredicate(
  (widget) =>
      widget is AppTextField && widget.field.label.startsWith('Pago de '),
);

String _textOf(WidgetTester tester, String label) =>
    tester.widget<AppTextField>(_field(label)).field.text.text;

String? _focusedLabel(WidgetTester tester) => tester
    .widgetList<AppTextField>(find.byType(AppTextField))
    .where((widget) => widget.field.focusNode.hasFocus)
    .firstOrNull
    ?.field
    .label;

Future<void> _type(WidgetTester tester, String label, String text) async {
  final field = find.descendant(
    of: _field(label),
    matching: find.byType(TextField),
  );
  await tester.ensureVisible(field);
  await tester.pumpAndSettle();
  await tester.tap(field);
  await tester.enterText(field, text);
  await tester.pump();
}

Future<void> _blur(WidgetTester tester) async {
  FocusManager.instance.primaryFocus?.unfocus();
  await tester.pumpAndSettle();
}

Future<void> _tapSave(WidgetTester tester) async {
  await tester.tap(find.widgetWithText(FilledButton, 'Guardar'));
  await tester.pump();
}

/// Opens the estimate, then the inputs form at [location] on top of it.
Future<void> _open(
  EstimateHarness harness,
  WidgetTester tester, {
  String? location,
  TaxInputs? inputs,
}) async {
  if (inputs != null) harness.repository.inputsFallback = inputs;
  await harness.pump(tester);
  await harness.push(tester, location ?? _inputs('612'));
}

TaxInputs _inputs601({List<Map<String, Object?>> values = const []}) =>
    taxInputsFrom(
      taxInputsJson(regime: '601', definitions: defs601, values: values),
    );

/// 612 for fiscal year 2026 seen from March 2027: all twelve months due,
/// January to [paidThrough] paid.
TaxInputs _pastYear612({int paidThrough = 1}) => taxInputsFrom(
  taxInputsJson(
    values: [
      for (var month = 1; month <= paidThrough; month++)
        valueJson(_priorPayments, month: month, cents: 100000),
    ],
  ),
);

void main() {
  late EstimateHarness harness;

  setUp(() => harness = EstimateHarness());

  testWidgets('reads the inputs of that regime and year and shows every '
      'definition, in order', (tester) async {
    await _open(harness, tester);

    expect(harness.repository.inputsCalls, [
      (taxpayerId: fixtureTaxpayerId, regime: '612', fiscalYear: 2026),
    ]);
    expect(find.text('Datos fiscales'), findsOneWidget);
    expect(find.text('Régimen 612 · Ejercicio 2026'), findsOneWidget);
    final tops = [
      for (final label in [
        'Pagos provisionales anteriores del ejercicio',
        'Ingresos del ejercicio antes de usar la app',
        'PTU pagada en el ejercicio',
      ])
        tester.getTopLeft(find.text(label, skipOffstage: false)).dy,
    ];
    expect(tops, orderedEquals([...tops]..sort()));
  });

  testWidgets('stored amounts show in es-MX; empty ones stay empty', (
    tester,
  ) async {
    await _open(harness, tester);

    expect(_textOf(tester, 'Pago de enero'), '2,000.00');
    expect(_textOf(tester, 'Pago de febrero'), isEmpty);
  });

  group('monthly prior payments', () {
    testWidgets('offer only the months already due this year', (tester) async {
      await _open(harness, tester);

      expect(_field('Pago de enero'), findsOneWidget);
      expect(_field('Pago de febrero'), findsOneWidget);
      expect(_field('Pago de marzo'), findsNothing);
      expect(find.text('Enero a junio'), findsNothing);
    });

    testWidgets("a past year's twelve months show six at a time", (
      tester,
    ) async {
      harness.now = DateTime.utc(2027, 3, 20, 18);
      await _open(harness, tester, inputs: _pastYear612());

      expect(_monthFields(), findsNWidgets(6));
      expect(_field('Pago de enero'), findsOneWidget);
      expect(
        tester
            .getSemantics(find.text('Julio a diciembre'))
            .flagsCollection
            .isExpanded,
        Tristate.isFalse,
      );

      await tester.ensureVisible(find.text('Julio a diciembre'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Julio a diciembre'));
      await tester.pumpAndSettle();

      expect(_monthFields(), findsNWidgets(6));
      expect(_field('Pago de julio'), findsOneWidget);
      expect(_field('Pago de enero'), findsNothing);
    });

    testWidgets('opens at the first month still missing', (tester) async {
      await _open(harness, tester);

      expect(_focusedLabel(tester), 'Pago de febrero');
    });

    testWidgets('a missing month in the second half opens that half', (
      tester,
    ) async {
      harness.now = DateTime.utc(2027, 3, 20, 18);
      await _open(harness, tester, inputs: _pastYear612(paidThrough: 7));

      expect(_focusedLabel(tester), 'Pago de agosto');
      expect(_field('Pago de agosto'), findsOneWidget);
      expect(_field('Pago de enero'), findsNothing);
    });
  });

  group("the app's own estimate as a suggestion (MODULE §8, OQ-T1)", () {
    const suggestion = r'Estimado por la app: $150.96 — confírmalo';

    testWidgets('an empty month offers its estimate, labelled, and saves '
        'nothing until it is used', (tester) async {
      await _open(harness, tester);

      expect(
        harness.repository.calls,
        contains((taxpayerId: fixtureTaxpayerId, period: '2026-02')),
      );
      expect(
        harness.repository.calls,
        isNot(contains((taxpayerId: fixtureTaxpayerId, period: '2026-01'))),
        reason: 'January holds what the user entered',
      );
      expect(find.text(suggestion), findsOneWidget);
      expect(_textOf(tester, 'Pago de febrero'), isEmpty);

      await _tapSave(tester);
      await tester.pumpAndSettle();

      expect(harness.repository.saves.single.values, const [
        TaxInputValue(key: _priorPayments, month: 1, valueCents: 200000),
      ]);
    });

    testWidgets('using it fills the month, which then saves', (tester) async {
      await _open(harness, tester);

      await tester.ensureVisible(find.text('Usar este monto'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Usar este monto'));
      await tester.pumpAndSettle();

      expect(_textOf(tester, 'Pago de febrero'), '150.96');
      expect(find.text(suggestion), findsNothing);
      await _tapSave(tester);
      await tester.pumpAndSettle();
      expect(harness.repository.saves.single.values, const [
        TaxInputValue(key: _priorPayments, month: 1, valueCents: 200000),
        TaxInputValue(key: _priorPayments, month: 2, valueCents: 15096),
      ]);
    });

    testWidgets('a month whose own estimate is incomplete offers nothing', (
      tester,
    ) async {
      await harness.pump(tester);
      harness.repository.fallback = estimateFrom(
        estimateJson(
          regimes: [regimeJson('612', status: 'incomplete', isrCents: 415096)],
        ),
      );
      await harness.push(tester, _inputs('612'));

      expect(find.textContaining('Estimado por la app'), findsNothing);
      expect(_field('Pago de febrero'), findsOneWidget);
    });

    testWidgets('a failed estimate offers nothing; the form still works', (
      tester,
    ) async {
      await harness.pump(tester);
      harness.repository.fallback = const ApiUnreachable();
      await harness.push(tester, _inputs('612'));

      expect(find.textContaining('Estimado por la app'), findsNothing);
      await _type(tester, 'Pago de febrero', '100');
      await _tapSave(tester);
      await tester.pumpAndSettle();
      expect(harness.repository.saves, hasLength(1));
    });

    testWidgets('meets contrast and target guidelines', (tester) async {
      await _open(harness, tester);
      await tester.ensureVisible(find.text('Usar este monto'));
      await tester.pumpAndSettle();

      await expectLater(tester, meetsGuideline(textContrastGuideline));
      await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
      await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
    });
  });

  testWidgets('opens at a one-amount field when the estimate asks for it', (
    tester,
  ) async {
    await _open(harness, tester, location: _inputs('612', field: 'ptu_paid'));

    expect(_focusedLabel(tester), 'PTU pagada en el ejercicio');
  });

  group('validation', () {
    testWidgets('a rate outside 0..1 is refused on blur', (tester) async {
      await _open(harness, tester, inputs: _inputs601());

      await _type(tester, 'Coeficiente de utilidad', '1.5');
      await _blur(tester);

      expect(find.text('Escribe un número entre 0 y 1.'), findsOneWidget);
    });

    testWidgets('a rate typed with a decimal comma is accepted', (
      tester,
    ) async {
      await _open(harness, tester, inputs: _inputs601());

      await _type(tester, 'Coeficiente de utilidad', '0,15');
      await _blur(tester);

      expect(_textOf(tester, 'Coeficiente de utilidad'), '0.15');
      expect(find.text('Escribe un número entre 0 y 1.'), findsNothing);
    });

    testWidgets('an amount that is not es-MX money is refused on blur', (
      tester,
    ) async {
      await _open(harness, tester);

      await _type(tester, 'PTU pagada en el ejercicio', '1.234,50');
      await _blur(tester);

      expect(
        find.text('Escribe un monto, por ejemplo 1,234.50.'),
        findsOneWidget,
      );
    });
  });

  testWidgets('saves every value in the field of its kind, then goes back '
      'to a refreshed estimate', (tester) async {
    await _open(
      harness,
      tester,
      location: _inputs('601', field: 'coeficiente_utilidad'),
      inputs: _inputs601(
        values: [valueJson(_priorPayments, month: 1, cents: 200000)],
      ),
    );
    final estimateReads = harness.repository.calls.length;

    await _type(tester, 'Coeficiente de utilidad', '0.15');
    await tester.ensureVisible(find.byKey(AppCheckboxField.rowKey));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(AppCheckboxField.rowKey));
    await _type(tester, 'Pago de febrero', '1,500');
    await _tapSave(tester);
    await tester.pumpAndSettle();

    expect(harness.repository.saves.single.regime, '601');
    expect(harness.repository.saves.single.fiscalYear, 2026);
    expect(harness.repository.saves.single.values, const [
      TaxInputValue(key: 'coeficiente_utilidad', valueRate: '0.15'),
      TaxInputValue(key: 'first_fiscal_year', valueBool: true),
      TaxInputValue(key: _priorPayments, month: 1, valueCents: 200000),
      TaxInputValue(key: _priorPayments, month: 2, valueCents: 150000),
    ]);
    expect(harness.location, TaxEngineRoutes.estimate);
    expect(harness.repository.calls.length, greaterThan(estimateReads));
    // History and Home below read again too (M8-07 device walk F2).
    expect(harness.container.read(taxDataRevisionProvider), 1);
    expect(find.textContaining('Guardamos tus datos'), findsOneWidget);
  });

  testWidgets('shows busy while saving and ignores repeated taps', (
    tester,
  ) async {
    final pending = Completer<TaxInputs>();
    harness.repository.saveReplies.add(pending);
    await _open(harness, tester);

    await _tapSave(tester);
    await _tapSave(tester);

    expect(find.bySemanticsLabel('Guardando…'), findsOneWidget);
    expect(harness.repository.saves, hasLength(1));
    pending.complete(taxInputsFrom(taxInputsJson()));
    await tester.pumpAndSettle();
    expect(harness.location, TaxEngineRoutes.estimate);
  });

  testWidgets('a read-only account sees its values; saving shows the '
      'paywall and keeps them', (tester) async {
    harness.repository.saveReplies.add(const ApiEntitlementRequired());
    await _open(harness, tester);

    expect(_textOf(tester, 'Pago de enero'), '2,000.00');
    await _type(tester, 'Pago de febrero', '500');
    await _tapSave(tester);
    await tester.pumpAndSettle();

    expect(find.text('Tu cuenta es de solo consulta'), findsOneWidget);
    expect(harness.container.read(readOnlyAccessProvider), isTrue);
    await tester.tap(find.text('Entendido'));
    await tester.pumpAndSettle();
    expect(harness.location, _inputs('612'));
    expect(_textOf(tester, 'Pago de febrero'), '500.00');
  });

  testWidgets("the server's verdict lands on the field it names", (
    tester,
  ) async {
    harness.repository.saveReplies.add(
      const ApiValidationFailure([
        FieldError(
          field: 'values.0.value_rate',
          code: 'out_of_range',
          message: 'x',
        ),
      ]),
    );
    await _open(
      harness,
      tester,
      location: _inputs('601', field: 'coeficiente_utilidad'),
      inputs: _inputs601(),
    );

    await _type(tester, 'Coeficiente de utilidad', '0.5');
    await _tapSave(tester);
    await tester.pumpAndSettle();

    expect(find.text('Revisa 1 dato antes de guardar'), findsOneWidget);
    expect(
      find.text('Escribe un número entre 0 y 1.'),
      findsAtLeastNWidgets(1),
    );
    expect(harness.location, contains('/datos/601'));
  });

  testWidgets('an archived taxpayer is told why nothing was saved', (
    tester,
  ) async {
    harness.repository.saveReplies.add(
      const ApiRejected(
        code: ErrorCode.businessRuleViolation,
        details: [
          FieldError(
            field: 'taxpayerId',
            code: 'taxpayer_archived',
            message: 'x',
          ),
        ],
      ),
    );
    await _open(harness, tester);

    await _tapSave(tester);
    await tester.pumpAndSettle();

    expect(find.textContaining('está archivado'), findsOneWidget);
  });

  testWidgets('a failed load offers a retry', (tester) async {
    harness.repository.inputsReplies.add(const ApiUnreachable());
    await _open(harness, tester);

    expect(find.textContaining('No hay conexión'), findsOneWidget);
    await tester.tap(find.text('Reintentar'));
    await tester.pumpAndSettle();

    expect(_field('Pago de enero'), findsOneWidget);
  });

  testWidgets('leaving with changes asks first', (tester) async {
    await _open(harness, tester);
    await _type(tester, 'PTU pagada en el ejercicio', '100');

    await tester.tap(find.byType(BackButton));
    await tester.pumpAndSettle();
    expect(find.text('¿Descartar los cambios?'), findsOneWidget);
    await tester.tap(find.text('Descartar'));
    await tester.pumpAndSettle();

    expect(harness.location, TaxEngineRoutes.estimate);
    expect(harness.repository.saves, isEmpty);
  });

  testWidgets('survives 200 % text on a phone', (tester) async {
    tester.view
      ..physicalSize = _phone
      ..devicePixelRatio = 1;
    addTearDown(tester.view.reset);
    tester.platformDispatcher.textScaleFactorTestValue = 2;
    addTearDown(tester.platformDispatcher.clearTextScaleFactorTestValue);
    harness.now = DateTime.utc(2027, 3, 20, 18);

    await _open(harness, tester, inputs: _pastYear612());
    await tester.ensureVisible(find.text('PTU pagada en el ejercicio'));
    await tester.pumpAndSettle();

    expect(tester.takeException(), isNull);
  });

  for (final brightness in Brightness.values) {
    testWidgets('meets contrast and target guidelines (${brightness.name})', (
      tester,
    ) async {
      harness.repository.inputsFallback = _inputs601();
      await harness.pump(
        tester,
        brightness: brightness,
        location: _inputs('601', field: 'coeficiente_utilidad'),
      );

      await expectLater(tester, meetsGuideline(textContrastGuideline));
      await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
      await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
    });
  }
}
