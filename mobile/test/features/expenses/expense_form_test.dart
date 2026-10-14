import 'dart:async';

import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/api/generated/models/payment_method.dart';
import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/api/generated/models/warning_code.dart';
import 'package:control_contable/core/forms/form_footer.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/features/expenses/expenses_routes.dart';
import 'package:control_contable/features/expenses/presentation/expense_category_field.dart';
import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';
import 'package:flutter_test/flutter_test.dart';

import '../taxpayers/pump_onboarding.dart' show inputLabeled, tapText, typeInto;
import '../taxpayers/taxpayers_fakes.dart';
import 'expenses_fakes.dart';
import 'pump_expenses.dart';

const _amount = 'Monto sin IVA';
const _paidOn = 'Fecha de pago';
const _ivaAmount = 'IVA pagado (opcional)';
const _uuid = 'Folio fiscal (UUID) (opcional)';
const _description = 'Descripción (opcional)';
const _isrWithheld = 'ISR retenido (opcional)';
const _ivaWithheld = 'IVA retenido (opcional)';
const _more = 'Más detalles';
const _withholdings = 'Retenciones a tu proveedor';
const _cashOverLimit =
    r'Los pagos en efectivo de más de $2,000.00 no son deducibles.';
const _cashIva = 'El IVA de este pago en efectivo no es acreditable.';

String _textOf(WidgetTester tester, String label) =>
    tester.widget<TextField>(inputLabeled(label)).controller!.text;

Future<void> _save(WidgetTester tester) => tapText(tester, 'Guardar');

/// Picks [name] in the category picker, searching for [query] first.
Future<void> pickCategory(
  WidgetTester tester,
  String name, {
  String? query,
}) async {
  final opener = find.descendant(
    of: find.byType(ExpenseCategoryField),
    matching: find.byType(OutlinedButton),
  );
  await tester.ensureVisible(opener.first);
  await tester.tap(opener.first);
  await tester.pumpAndSettle();
  if (query != null) {
    await tester.enterText(
      find.widgetWithText(TextField, 'Buscar categoría'),
      query,
    );
    await tester.pumpAndSettle();
  }
  await tester.tap(find.text(name).last);
  await tester.pumpAndSettle();
}

Future<void> _toggleDeductible(WidgetTester tester) async {
  await tester.ensureVisible(find.byType(SwitchListTile));
  await tester.pumpAndSettle();
  await tester.tap(find.byType(SwitchListTile));
  await tester.pumpAndSettle();
}

SwitchListTile _deductibleSwitch(WidgetTester tester) =>
    tester.widget<SwitchListTile>(find.byType(SwitchListTile));

void main() {
  late ExpensesHarness harness;

  setUp(() => harness = ExpensesHarness());

  Future<void> pumpForm(
    WidgetTester tester, {
    Taxpayer? taxpayer,
    Brightness brightness = Brightness.light,
  }) async {
    await harness.pump(
      tester,
      brightness: brightness,
      taxpayer: taxpayer ?? anaOneRegime,
    );
    unawaited(
      harness.container
          .read(appRouterProvider)
          .push<void>(ExpensesRoutes.create),
    );
    await tester.pumpAndSettle();
  }

  /// 500.00 of software, the defaults otherwise (US-E1).
  Future<void> fillSoftware(WidgetTester tester) async {
    await pickCategory(tester, 'Software y suscripciones');
    await typeInto(tester, _amount, '500');
  }

  group('M4-10 layout and defaults', () {
    // QA M3/M4 D1: the picker is a button for assistive technology too.
    // TalkBack and VoiceOver activate a node through its tap action; the
    // category's merged node had the button role but no action, so the
    // picker could not be opened without touch.
    testWidgets('the category picker opens from its semantics tap action', (
      tester,
    ) async {
      final handle = tester.ensureSemantics();
      await pumpForm(tester);

      final category = find.semantics.byLabel(RegExp('^Categoría: '));
      expect(category, findsOne);
      tester.semantics.tap(category);
      await tester.pumpAndSettle();

      expect(find.text('Elige la categoría del gasto'), findsOneWidget);
      handle.dispose();
    });

    testWidgets('a page with the category first, the legend under it, then '
        'amount, IVA, date, payment and CFDI; defaults prefilled; the rest '
        'collapsed', (tester) async {
      await pumpForm(tester);

      expect(find.text('Nuevo gasto'), findsOneWidget);
      expect(harness.expenses.categoryRequests.last, (
        regime: '612',
        on: '2026-10-13',
      ));
      final order = [
        'Categoría',
        expenseLegend,
        _amount,
        'Tasa de IVA',
        _paidOn,
        'Forma de pago',
        '¿Tienes factura?',
      ];
      for (var i = 1; i < order.length; i++) {
        expect(
          tester.getTopLeft(find.text(order[i - 1]).first).dy,
          lessThan(tester.getTopLeft(find.text(order[i]).first).dy),
          reason: '${order[i - 1]} before ${order[i]}',
        );
      }
      expect(_textOf(tester, _paidOn), '13/10/2026');
      RadioListTile<String> radio(String label) =>
          tester.widget<RadioListTile<String>>(
            find.widgetWithText(RadioListTile<String>, label),
          );
      final group = tester.widget<RadioGroup<String>>(
        find.ancestor(
          of: find.widgetWithText(RadioListTile<String>, '16 %'),
          matching: find.byType(RadioGroup<String>),
        ),
      );
      expect(group.groupValue, IvaRate.rate16.json);
      expect(radio('Tarjeta de débito').value, PaymentMethod.cardDebit.json);
      expect(find.text('Régimen'), findsNothing);
      expect(find.text(_uuid), findsNothing);
      expect(find.text(_more), findsOneWidget);
      // A persona física withholds nothing from suppliers.
      expect(find.text(_withholdings), findsNothing);
    });

    testWidgets('the picker searches the categories; the choice shows its '
        'plain-language hint, note and basis above the legend', (tester) async {
      await pumpForm(tester);

      await pickCategory(tester, 'Software y suscripciones', query: 'soft');
      expect(find.text('Normalmente deducible para tu régimen.'), findsOne);
      expect(find.text('Fundamento: LISR 103'), findsOneWidget);

      await pickCategory(tester, 'Consumo en restaurantes', query: 'restau');
      expect(find.text('Deducible sólo en parte para tu régimen.'), findsOne);
      expect(find.text('Sólo una parte es deducible'), findsOneWidget);
      expect(
        tester.getTopLeft(find.text('Sólo una parte es deducible')).dy,
        lessThan(tester.getTopLeft(find.text(expenseLegend)).dy),
      );
    });

    testWidgets('RESICO PF: the hint says expenses never lower ISR but the '
        'IVA may be credited', (tester) async {
      await pumpForm(tester, taxpayer: anaResico);

      await pickCategory(tester, 'Software y suscripciones');

      expect(find.text('No reduce tu ISR en RESICO.'), findsOneWidget);
      expect(
        find.text(
          'En RESICO tus gastos no reducen tu ISR, pero el IVA que pagaste '
          'sí puede acreditarse.',
        ),
        findsOneWidget,
      );
    });

    testWidgets('the hint of each chosen category is announced', (
      tester,
    ) async {
      await pumpForm(tester);
      tester.takeAnnouncements();

      await pickCategory(tester, 'Software y suscripciones', query: 'soft');
      expect(
        tester.takeAnnouncements().map((a) => a.message),
        contains('Normalmente deducible para tu régimen.'),
      );

      await pickCategory(tester, 'Consumo en restaurantes', query: 'restau');
      expect(
        tester.takeAnnouncements().map((a) => a.message),
        contains('Deducible sólo en parte para tu régimen.'),
      );
    });

    testWidgets('the total with IVA follows the amount live (integer '
        'centavos)', (tester) async {
      await pumpForm(tester);

      await typeInto(tester, _amount, '1,234.56');

      expect(find.text(r'Total con IVA $1,432.09'), findsOneWidget);
    });

    testWidgets('a failed category load offers a retry', (tester) async {
      harness.expenses.failNext['categories'] = const ApiUnreachable();
      await pumpForm(tester);

      expect(find.text('No se pudieron cargar las categorías.'), findsOne);
      await tapText(tester, 'Reintentar');
      expect(find.text(expenseLegend), findsOneWidget);
    });
  });

  group('M4-10 cash payments', () {
    testWidgets('cash above the limit warns at once, inline and never '
        'blocking: not deductible, IVA not creditable', (tester) async {
      await pumpForm(tester);
      await pickCategory(tester, 'Papelería y artículos de oficina');
      await typeInto(tester, _amount, '1000');
      await tapText(tester, 'Efectivo');

      expect(find.text(_cashOverLimit), findsNothing);

      await typeInto(tester, _amount, '2000');

      expect(find.text(_cashOverLimit), findsOneWidget);
      expect(find.text(_cashIva), findsOneWidget);
      await _save(tester);
      expect(harness.expenses.creates.single.paymentMethod, PaymentMethod.cash);
    });

    testWidgets('a cash warning is announced once when it appears (a live '
        'region alone is silent on VoiceOver)', (tester) async {
      await pumpForm(tester);
      await pickCategory(tester, 'Papelería y artículos de oficina');
      await typeInto(tester, _amount, '1000');
      await tapText(tester, 'Efectivo');
      tester.takeAnnouncements();

      await typeInto(tester, _amount, '2000');

      final announced = tester.takeAnnouncements();
      expect(announced.map((a) => a.message), [_cashOverLimit, _cashIva]);
      expect(
        announced.map((a) => a.assertiveness),
        everyElement(Assertiveness.polite),
      );

      await typeInto(tester, _amount, '2500');

      expect(tester.takeAnnouncements(), isEmpty);
    });

    testWidgets('fuel paid in cash warns whatever the amount', (tester) async {
      await pumpForm(tester);
      await pickCategory(tester, 'Combustibles');
      await typeInto(tester, _amount, '300');
      await tapText(tester, 'Efectivo');

      expect(
        find.text('El combustible pagado en efectivo no es deducible.'),
        findsOneWidget,
      );
      expect(find.text(_cashIva), findsOneWidget);
    });

    testWidgets('the suggested deductible switch follows the cash rule', (
      tester,
    ) async {
      await pumpForm(tester);
      await fillSoftware(tester);
      await tapText(tester, _more);
      expect(_deductibleSwitch(tester).value, isTrue);

      await typeInto(tester, _amount, '5000');
      await tapText(tester, 'Efectivo');

      expect(_deductibleSwitch(tester).value, isFalse);
    });
  });

  group('M4-10 save', () {
    testWidgets('Guardar sends the expense with the session id and the '
        'default deductibility, then goes back to the list (US-E1)', (
      tester,
    ) async {
      await pumpForm(tester);
      await fillSoftware(tester);

      await _save(tester);

      final body = harness.expenses.creates.single;
      expect(body.id, harness.clientIds.first);
      expect(body.regimeCode, '612');
      expect(body.categoryCode, 'software');
      expect(body.amountCents, 50000);
      expect(body.ivaRate, IvaRate.rate16);
      expect(body.ivaCents, 8000);
      expect(body.paymentMethod, PaymentMethod.cardDebit);
      expect(body.paidOn, '2026-10-13');
      expect(body.hasCfdi, isTrue);
      expect(body.deductible, isNull);
      expect(body.cfdiUuid, isNull);
      expect(body.isrWithheldCents, isNull);
      expect(harness.location, ExpensesRoutes.list);
      expect(find.text('Gasto guardado'), findsOneWidget);
    });

    testWidgets('an empty submit lists what is missing and sends nothing', (
      tester,
    ) async {
      await pumpForm(tester);

      await _save(tester);

      expect(harness.expenses.creates, isEmpty);
      expect(find.text('Elige una opción.'), findsWidgets);
    });

    testWidgets('a retry after a lost answer resends the same id', (
      tester,
    ) async {
      harness.expenses.failNext['create'] = const ApiUnreachable();
      await pumpForm(tester);
      await fillSoftware(tester);

      await _save(tester);
      expect(
        find.textContaining('No se pudo guardar por falta de conexión'),
        findsOneWidget,
      );
      await _save(tester);

      expect(harness.expenses.creates, hasLength(2));
      expect(harness.expenses.creates.map((c) => c.id).toSet(), {
        harness.clientIds.first,
      });
    });

    testWidgets('no CFDI hides the folio, warns inline and is sent', (
      tester,
    ) async {
      await pumpForm(tester);
      await fillSoftware(tester);
      await tapText(tester, 'No');
      await tapText(tester, _more);

      expect(find.text(_uuid), findsNothing);
      expect(find.text('Sin factura, el gasto no es deducible.'), findsOne);
      await _save(tester);
      expect(harness.expenses.creates.single.hasCfdi, isFalse);
    });

    testWidgets('the override switch sends the user\'s choice; "Usar el '
        'valor sugerido" clears it', (tester) async {
      await pumpForm(tester);
      await pickCategory(tester, 'Consumo en restaurantes');
      await typeInto(tester, _amount, '800');
      await tapText(tester, _more);
      expect(_deductibleSwitch(tester).value, isFalse);

      await _toggleDeductible(tester);
      expect(_deductibleSwitch(tester).value, isTrue);
      await tapText(tester, 'Usar el valor sugerido');
      expect(_deductibleSwitch(tester).value, isFalse);
      await _toggleDeductible(tester);

      await _save(tester);
      expect(harness.expenses.creates.single.deductible, isTrue);
    });

    testWidgets('Guardar y crear otro keeps date, payment method and CFDI, '
        'clears the rest, starts a new id and focuses the category', (
      tester,
    ) async {
      await pumpForm(tester);
      await fillSoftware(tester);
      await tapText(tester, 'Transferencia');
      await typeInto(tester, _paidOn, '10/10/2026');

      await tapText(tester, 'Guardar y crear otro');

      expect(harness.location, ExpensesRoutes.create);
      expect(find.text('Gasto guardado'), findsOneWidget);
      expect(_textOf(tester, _amount), isEmpty);
      expect(_textOf(tester, _paidOn), '10/10/2026');
      expect(find.text('Elige una categoría'), findsOneWidget);
      final categoryButton = tester.widget<OutlinedButton>(
        find.descendant(
          of: find.byType(ExpenseCategoryField),
          matching: find.byType(OutlinedButton),
        ),
      );
      expect(categoryButton.focusNode?.hasPrimaryFocus, isTrue);

      await fillSoftware(tester);
      await _save(tester);
      final second = harness.expenses.creates.last;
      expect(second.id, harness.clientIds[1]);
      expect(second.paymentMethod, PaymentMethod.transfer);
      expect(second.paidOn, '2026-10-10');
      // The next form starts from the last payment method used.
      unawaited(
        harness.container
            .read(appRouterProvider)
            .push<void>(ExpensesRoutes.create),
      );
      await tester.pumpAndSettle();
      final group = tester.widget<RadioGroup<String>>(
        find.ancestor(
          of: find.widgetWithText(RadioListTile<String>, 'Efectivo'),
          matching: find.byType(RadioGroup<String>),
        ),
      );
      expect(group.groupValue, PaymentMethod.transfer.json);
    });

    testWidgets('the warnings of the answer stay in a toast until closed, '
        'provisional parameters included', (tester) async {
      harness.expenses.warnNext = const [
        Warning(
          code: WarningCode.provisionalParameters,
          field: 'paid_on',
          message: 'x',
        ),
      ];
      await pumpForm(tester);
      await fillSoftware(tester);

      await _save(tester);

      expect(
        find.textContaining(
          'Se usaron los parámetros fiscales del año anterior',
        ),
        findsOneWidget,
      );
      expect(tester.widget<SnackBar>(find.byType(SnackBar)).persist, isTrue);
    });

    // QA M3/M4 D3: the focused first field is shown, not left above the
    // fold where the previous record was finished (WCAG 2.2 · 2.4.11).
    testWidgets('Guardar y crear otro brings the focused category into view', (
      tester,
    ) async {
      await pumpForm(tester);
      await fillSoftware(tester);
      await tester.drag(
        find.byType(SingleChildScrollView).first,
        const Offset(0, -2000),
      );
      await tester.pumpAndSettle();
      final opener = find.descendant(
        of: find.byType(ExpenseCategoryField),
        matching: find.byType(OutlinedButton),
      );
      final viewport = tester.getRect(find.byType(SingleChildScrollView).first);
      expect(
        tester.getRect(opener).top,
        lessThan(viewport.top),
        reason: 'scrolled away first',
      );

      await tapText(tester, 'Guardar y crear otro');

      final shown = tester.getRect(opener);
      expect(shown.top, greaterThanOrEqualTo(viewport.top));
      expect(shown.bottom, lessThanOrEqualTo(viewport.bottom));
      expect(
        tester.widget<OutlinedButton>(opener).focusNode?.hasPrimaryFocus,
        isTrue,
      );
    });

    // QA M3/M4 D2: after *Guardar y crear otro* the warning toast stays
    // until closed, on the form itself; it must not cover the footer.
    testWidgets('a persistent warning toast leaves the form footer clear', (
      tester,
    ) async {
      harness.expenses.warnNext = const [
        Warning(code: WarningCode.noCfdi, field: 'has_cfdi', message: 'x'),
      ];
      await pumpForm(tester);
      await fillSoftware(tester);

      await tapText(tester, 'Guardar y crear otro');

      // The visible surface: a floating SnackBar's own box includes its margin.
      final toast = tester.getRect(
        find
            .descendant(
              of: find.byType(SnackBar),
              matching: find.byType(Material),
            )
            .first,
      );
      final footer = tester.getRect(find.byType(FormFooter));
      expect(toast.bottom, lessThanOrEqualTo(footer.top));
    });

    testWidgets('an unknown category lands on the category field', (
      tester,
    ) async {
      harness.expenses.failNext['create'] = ApiValidationFailure([
        detail('category_code', 'unknown_category'),
      ]);
      await pumpForm(tester);
      await fillSoftware(tester);

      await _save(tester);

      expect(find.text('Elige una categoría de la lista.'), findsWidgets);
      expect(harness.location, ExpensesRoutes.create);
    });

    testWidgets('a spent trial cap names the expenses cap and keeps the '
        'form', (tester) async {
      harness.expenses.failNext['create'] = ApiEntitlementRequired(
        details: [detail('expenses_remaining', 'trial_cap_reached')],
      );
      await pumpForm(tester);
      await fillSoftware(tester);

      await _save(tester);

      expect(find.text('Llegaste al límite de gastos de prueba'), findsOne);
      expect(find.textContaining('5 gastos'), findsWidgets);
      expect(harness.location, ExpensesRoutes.create);
    });

    testWidgets('a read-only account gets the read-only paywall', (
      tester,
    ) async {
      harness.expenses.failNext['create'] = const ApiEntitlementRequired();
      await pumpForm(tester);
      await fillSoftware(tester);

      await _save(tester);

      expect(find.text('Tu cuenta es de solo consulta'), findsOneWidget);
    });
  });

  group('M4-10 persona moral withholdings', () {
    testWidgets('professional fees open the supplier withholdings, which '
        'are sent', (tester) async {
      await pumpForm(tester, taxpayer: moral601);
      expect(find.text(_withholdings), findsOneWidget);
      expect(find.text(_isrWithheld), findsNothing);

      await pickCategory(tester, 'Honorarios y servicios profesionales');
      await typeInto(tester, _amount, '10000');
      await typeInto(tester, _isrWithheld, '1000');
      await typeInto(tester, _ivaWithheld, '1066.67');
      await _save(tester);

      final body = harness.expenses.creates.single;
      expect(body.regimeCode, '601');
      expect(body.isrWithheldCents, 100000);
      expect(body.ivaWithheldCents, 106667);
    });
  });

  group('M4-10 more details', () {
    testWidgets('IVA, folio and description travel when written', (
      tester,
    ) async {
      await pumpForm(tester);
      await fillSoftware(tester);
      await tapText(tester, _more);
      expect(_textOf(tester, _ivaAmount), '80.00');

      await typeInto(tester, _ivaAmount, '79.50');
      await typeInto(tester, _uuid, '3a9c7e21-5b4d-4f6a-8e2c-9d1b7a3f5c60');
      await typeInto(tester, _description, 'Suscripción anual');
      await _save(tester);

      final body = harness.expenses.creates.single;
      expect(body.ivaCents, 7950);
      expect(body.cfdiUuid, '3A9C7E21-5B4D-4F6A-8E2C-9D1B7A3F5C60');
      expect(body.description, 'Suscripción anual');
    });
  });

  group('M4-10 flag', () {
    testWidgets('module_expenses off: the form is not found', (tester) async {
      await harness.pump(tester, enabled: false, location: null);

      unawaited(
        harness.container
            .read(appRouterProvider)
            .push<void>(ExpensesRoutes.create),
      );
      await tester.pumpAndSettle();

      expect(harness.location, AppRoutes.notFound);
      expect(harness.expenses.calls, isEmpty);
    });
  });

  for (final brightness in Brightness.values) {
    testWidgets('meets the tap-target, label and contrast guidelines '
        '(${brightness.name})', (tester) async {
      final handle = tester.ensureSemantics();
      await pumpForm(tester, brightness: brightness);
      await fillSoftware(tester);
      await tapText(tester, _more);

      await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
      await expectLater(tester, meetsGuideline(iOSTapTargetGuideline));
      await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
      await expectLater(tester, meetsGuideline(textContrastGuideline));
      handle.dispose();
    });
  }

  testWidgets('leaving with unsaved changes asks before discarding', (
    tester,
  ) async {
    await pumpForm(tester);
    await typeInto(tester, _amount, '10');

    await tapText(tester, 'Cancelar');

    expect(find.text('¿Descartar este gasto?'), findsOneWidget);
    await tapText(tester, 'Descartar');
    expect(harness.location, ExpensesRoutes.list);
  });
}
