import 'dart:async';

import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/api/generated/models/expense_category.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/api/generated/models/operation_kind.dart';
import 'package:control_contable/api/generated/models/supplier_kind.dart';
import 'package:control_contable/api/generated/models/supplier_regime.dart';
import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/api/generated/models/warning_code.dart';
import 'package:control_contable/api/generated/models/withholder.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/features/expenses/expenses_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/choice_finders.dart';
import '../../helpers/withholding_fixtures.dart';
import '../taxpayers/pump_onboarding.dart' show tapText, typeInto;
import '../taxpayers/taxpayers_fakes.dart';
import 'expenses_fakes.dart' show businessCategories, expense;
import 'pump_expenses.dart';

const _amount = 'Monto sin IVA';
const _paidOn = 'Fecha de pago';
const _ivaRate = 'Tasa de IVA';
const _section = 'Tu proveedor';
const _e1 = '¿Quién te vendió o te dio el servicio?';
const _e2 = '¿En qué régimen está tu proveedor?';
const _e3 = '¿Qué te vendió o qué servicio te dio?';
const _e4 = '¿Es una casa o departamento para vivir?';
const _pm = 'Una empresa (persona moral)';
const _pf = 'Una persona física con negocio o profesión';
const _foreign = 'Alguien en el extranjero';
const _fees = 'Honorarios y servicios profesionales';
const _software = 'Software y suscripciones';
const _rentCategory = 'Renta de local, oficina o bodega';
const _foreignWarning =
    'Las operaciones con el extranjero pueden llevar otros impuestos que la '
    'app todavía no calcula.';

/// The business catalog with software served with [operation] as its
/// default (`null`: the server sends none yet).
List<ExpenseCategory> _softwareServedAs(OperationKind? operation) => [
  for (final c in businessCategories())
    if (c.code == 'software')
      c.copyWith(defaultOperationKind: operation)
    else
      c,
];

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
      taxpayer: taxpayer ?? moral601,
    );
    unawaited(
      harness.container
          .read(appRouterProvider)
          .push<void>(ExpensesRoutes.create),
    );
    await tester.pumpAndSettle();
  }

  group('WH-12 a persona física is never asked (US-W7)', () {
    testWidgets('no "Tu proveedor", no card, no supplier fact sent', (
      tester,
    ) async {
      await pumpForm(tester, taxpayer: anaOneRegime);
      await pickCategory(tester, _fees);
      await typeInto(tester, _amount, '1000');
      await settlePreview(tester);

      expect(find.text(_section), findsNothing);
      expect(find.text(_e1), findsNothing);
      expect(find.byKey(const ValueKey('withholding-card')), findsNothing);
      expect(harness.expenses.previewRequests, isEmpty);
      await tapText(tester, 'Guardar');
      final sent = harness.expenses.creates.single;
      expect(sent.supplierKind, isNull);
      expect(sent.supplierRegime, isNull);
      expect(sent.operationKind, isNull);
      expect(sent.residentialRent, isNull);
    });

    testWidgets('a supplier_not_applicable verdict says so in plain words', (
      tester,
    ) async {
      harness.expenses.failNext['create'] = ApiRejected(
        code: ErrorCode.businessRuleViolation,
        details: [detail('supplier_kind', 'supplier_not_applicable')],
      );
      await pumpForm(tester, taxpayer: anaOneRegime);
      await pickCategory(tester, _software);
      await typeInto(tester, _amount, '1000');

      await tapText(tester, 'Guardar');

      expect(
        find.textContaining('Una persona física no retiene a sus proveedores'),
        findsOneWidget,
      );
    });
  });

  group('WH-12 "Tu proveedor" for a persona moral', () {
    testWidgets('comes right after the category; an ordinary purchase '
        'starts collapsed at "una empresa"', (tester) async {
      await pumpForm(tester);
      await pickCategory(tester, _software);

      expect(
        tester.getTopLeft(find.text(_section)).dy,
        lessThan(tester.getTopLeft(find.text(_amount)).dy),
      );
      expect(find.text(_e1), findsNothing);
      await tapText(tester, _section);
      expect(chosenIn(tester, _e1), SupplierKind.pm.json);
      expect(chosenIn(tester, _e3), OperationKind.businessServices.json);
      expect(find.text(_e2), findsNothing);
      expect(find.text(_e4), findsNothing);

      await typeInto(tester, _amount, '1000');
      await tapText(tester, 'Guardar');
      final sent = harness.expenses.creates.single;
      expect(sent.supplierKind, SupplierKind.pm);
      expect(sent.operationKind, OperationKind.businessServices);
      expect(sent.supplierRegime, isNull);
    });

    testWidgets('fees open it with no preselection, and an unanswered '
        'supplier blocks only on submit', (tester) async {
      await pumpForm(tester);
      await pickCategory(tester, _fees);

      expect(find.text(_e1), findsOneWidget);
      expect(chosenIn(tester, _e1), isNull);
      // What was sold is asked once the supplier is known.
      expect(find.text(_e3), findsNothing);
      await typeInto(tester, _amount, '1000');

      await tapText(tester, 'Guardar');

      expect(find.text('Elige una opción.'), findsOneWidget);
      expect(harness.expenses.creates, isEmpty);
    });

    testWidgets('the regime is asked only of a persona física supplier, '
        '"No sé" by default', (tester) async {
      await pumpForm(tester);
      await pickCategory(tester, _fees);
      expect(find.text(_e2), findsNothing);

      await chooseIn(tester, _e1, _pf);

      expect(find.text(_e2), findsOneWidget);
      expect(find.text('Viene en su factura, junto a su RFC.'), findsOneWidget);
      expect(chosenIn(tester, _e2), SupplierRegime.unknown.json);
      expect(chosenIn(tester, _e3), OperationKind.professionalServices.json);
      await chooseIn(tester, _e1, _pm);
      expect(find.text(_e2), findsNothing);
    });

    testWidgets('the operation follows the category until the user picks '
        'one; rent asks if it is housing, and "Sí" makes it exempt', (
      tester,
    ) async {
      await pumpForm(tester);
      await pickCategory(tester, _rentCategory);
      await chooseIn(tester, _e1, _pm);

      expect(chosenIn(tester, _e3), OperationKind.rentRealEstate.json);
      expect(find.text(_e4), findsOneWidget);
      await typeInto(tester, _amount, '30000');
      await chooseIn(tester, _e4, 'Sí');

      expect(find.text(_ivaRate), findsNothing);
      await tapText(tester, 'Guardar');
      final sent = harness.expenses.creates.single;
      expect(sent.operationKind, OperationKind.rentRealEstate);
      expect(sent.residentialRent, isTrue);
      expect(sent.ivaRate, IvaRate.exempt);
      expect(sent.ivaCents, 0);
    });

    testWidgets('the operation starts at the default the catalog serves '
        '(default_operation_kind), never at a table of the app', (
      tester,
    ) async {
      harness.expenses.categoriesBy['601'] = _softwareServedAs(
        OperationKind.commission,
      );
      await pumpForm(tester);
      await pickCategory(tester, _software);
      await tapText(tester, _section);

      expect(chosenIn(tester, _e3), OperationKind.commission.json);
      await typeInto(tester, _amount, '1000');
      await settlePreview(tester);
      expect(
        harness.expenses.previewRequests.last.operationKind,
        OperationKind.commission,
      );
      await tapText(tester, 'Guardar');
      expect(
        harness.expenses.creates.single.operationKind,
        OperationKind.commission,
      );
    });

    testWidgets('a category served without a default leaves the operation '
        'unanswered and omits it, so the server applies its own', (
      tester,
    ) async {
      harness.expenses.categoriesBy['601'] = _softwareServedAs(null);
      await pumpForm(tester);
      await pickCategory(tester, _software);
      await tapText(tester, _section);

      expect(find.text(_e3), findsOneWidget);
      expect(chosenIn(tester, _e3), isNull);
      await typeInto(tester, _amount, '1000');
      await settlePreview(tester);
      expect(harness.expenses.previewRequests.last.operationKind, isNull);
      await tapText(tester, 'Guardar');
      expect(harness.expenses.creates.single.operationKind, isNull);
    });

    testWidgets('answers the user chose survive a change of category; only '
        'suggestions follow it', (tester) async {
      await pumpForm(tester);
      await pickCategory(tester, _fees);
      await chooseIn(tester, _e1, _pf);
      await chooseIn(tester, _e2, 'RESICO (626)');

      await pickCategory(tester, _rentCategory);

      expect(chosenIn(tester, _e1), SupplierKind.pfWithActivity.json);
      expect(chosenIn(tester, _e2), SupplierRegime.value626.json);
      // Not chosen by the user: it follows the new category.
      expect(chosenIn(tester, _e3), OperationKind.rentRealEstate.json);
    });

    testWidgets('abroad: nothing else is asked, and it warns', (tester) async {
      await pumpForm(tester);
      await pickCategory(tester, _fees);

      await chooseIn(tester, _e1, _foreign);

      expect(find.text(_e2), findsNothing);
      expect(find.text(_e3), findsNothing);
      expect(find.text(_foreignWarning), findsOneWidget);
      await typeInto(tester, _amount, '1000');
      await tapText(tester, 'Guardar');
      final sent = harness.expenses.creates.single;
      expect(sent.supplierKind, SupplierKind.foreign);
      expect(sent.operationKind, isNull);
    });

    testWidgets('the next expense of the same category starts at the last '
        'supplier used for it', (tester) async {
      await pumpForm(tester);
      await pickCategory(tester, _fees);
      await chooseIn(tester, _e1, _pf);
      await chooseIn(tester, _e2, 'RESICO (626)');
      await typeInto(tester, _amount, '1000');
      await tapText(tester, 'Guardar y crear otro');

      await pickCategory(tester, _fees);

      expect(chosenIn(tester, _e1), SupplierKind.pfWithActivity.json);
      expect(chosenIn(tester, _e2), SupplierRegime.value626.json);
    });

    // WH-13 QA O2: an expense stored before WH-07 (not backfilled yet) has
    // no operation_kind; the server derives it with the category's default
    // when it is omitted, so E3 must show that default instead of an
    // unanswered question next to a card that already used it.
    testWidgets('a legacy expense without an operation shows the category '
        'default the server applies', (tester) async {
      harness.expenses.categoriesBy['601'] = _softwareServedAs(
        OperationKind.commission,
      );
      harness.expenses.expenses.add(expense(id: 'e-legacy', regime: '601'));
      await harness.pump(tester, taxpayer: moral601);
      await tester.tap(find.text('Suscripción de diseño'));
      await tester.pumpAndSettle();

      await chooseIn(tester, _e1, _pm);

      expect(chosenIn(tester, _e3), OperationKind.commission.json);
      await settlePreview(tester);
      expect(
        harness.expenses.previewRequests.last.operationKind,
        OperationKind.commission,
      );
    });
  });

  group('WH-12 the derived card', () {
    testWidgets('US-W6: a 601 pays a 612 consultant 40,000.00 + IVA ⇒ '
        'withhold 4,000.00 and 4,266.67, pay 38,133.33, remit 8,266.67 by '
        'July 17, deduction conditioned', (tester) async {
      harness.expenses.previewAnswer = (_) => moralPaysConsultant;
      await pumpForm(tester);
      await typeInto(tester, _paidOn, '15062026');
      await pickCategory(tester, _fees);
      await chooseIn(tester, _e1, _pf);
      await chooseIn(tester, _e2, 'Actividad empresarial o profesional (612)');
      await typeInto(tester, _amount, '40000');
      await settlePreview(tester);

      final query = harness.expenses.previewRequests.last;
      expect(query.regime, '601');
      expect(query.on, '2026-06-15');
      expect(query.categoryCode, 'servicios_profesionales');
      expect(query.amountCents, 4000000);
      expect(query.ivaCents, 640000);
      expect(query.supplierKind, SupplierKind.pfWithActivity);
      expect(query.supplierRegime, SupplierRegime.value612);
      expect(query.operationKind, OperationKind.professionalServices);
      expect(
        find.text(r'Debes retener: ISR $4,000.00 · IVA $4,266.67'),
        findsOneWidget,
      );
      expect(find.text(r'Pagas a tu proveedor $38,133.33'), findsOneWidget);
      expect(
        find.text(r'Entregas al SAT $8,266.67 a más tardar el 17 de julio.'),
        findsOneWidget,
      );
      expect(
        find.text(
          'Este gasto solo es deducible si enteras las retenciones a más '
          'tardar el 17 de julio.',
        ),
        findsOneWidget,
      );
    });

    // iOS QA walk 2026-10-15: with the amount typed first, changing the
    // payment date afterwards kept the card on today's deadline ("17 de
    // noviembre") although the server derives July 17 for a June payment.
    testWidgets('changing the payment date after the amount asks again with '
        'the new date', (tester) async {
      harness.expenses.previewAnswer = (_) => moralPaysConsultant;
      await pumpForm(tester);
      await pickCategory(tester, _fees);
      await chooseIn(tester, _e1, _pf);
      await chooseIn(tester, _e2, 'Actividad empresarial o profesional (612)');
      await typeInto(tester, _amount, '40000');
      await settlePreview(tester);
      final before = harness.expenses.previewRequests.length;

      await typeInto(tester, _paidOn, '15062026');
      await settlePreview(tester);

      expect(harness.expenses.previewRequests.length, greaterThan(before));
      expect(harness.expenses.previewRequests.last.on, '2026-06-15');
    });

    testWidgets('an unknown supplier regime warns that it changes the '
        'withholding', (tester) async {
      harness.expenses.previewAnswer = (_) => preview(
        moralPaysConsultant.withholding,
        warnings: const [
          WarningCode.supplierRegimeUnknown,
          WarningCode.deductionRequiresRemittance,
        ],
      );
      await pumpForm(tester);
      await pickCategory(tester, _fees);
      await chooseIn(tester, _e1, _pf);
      await typeInto(tester, _amount, '40000');
      await settlePreview(tester);

      expect(
        find.text(
          'Revisa el régimen en la factura de tu proveedor: cambia la '
          'retención.',
        ),
        findsOneWidget,
      );
    });

    testWidgets('a company supplier of ordinary goods: nothing to withhold', (
      tester,
    ) async {
      await pumpForm(tester);
      await pickCategory(tester, _software);
      await typeInto(tester, _amount, '1000');
      await settlePreview(tester);

      expect(
        find.text('No debes retener impuestos en esta compra.'),
        findsOneWidget,
      );
      expect(find.text(r'Pagas a tu proveedor $1,160.00'), findsOneWidget);
    });

    testWidgets('the derived amounts are announced politely', (tester) async {
      harness.expenses.previewAnswer = (_) => preview(
        withholding(
          isrCents: 400000,
          ivaCents: 426667,
          withholder: Withholder.you,
          netCents: 3813333,
          remitDueOn: '2026-11-17',
        ),
      );
      await pumpForm(tester);
      await pickCategory(tester, _fees);
      await chooseIn(tester, _e1, _pf);
      tester.takeAnnouncements();

      await typeInto(tester, _amount, '40000');
      await settlePreview(tester);

      final announced = tester.takeAnnouncements();
      expect(
        announced.map((a) => a.message),
        contains(
          r'Debes retener: ISR $4,000.00 · IVA $4,266.67. '
          r'Pagas a tu proveedor $38,133.33',
        ),
      );
      expect(
        announced.map((a) => a.assertiveness),
        everyElement(Assertiveness.polite),
      );
    });
  });

  for (final brightness in Brightness.values) {
    testWidgets('"Tu proveedor" and the card meet the tap-target, label and '
        'contrast guidelines, and survive 200 % text (${brightness.name})', (
      tester,
    ) async {
      // Tall enough that no text sits under the sticky footer (the contrast
      // check reads the pixels on screen).
      tester.view.physicalSize = const Size(1080, 6000);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.reset);
      final handle = tester.ensureSemantics();
      harness.expenses.previewAnswer = (_) => moralPaysConsultant;
      await pumpForm(tester, brightness: brightness);
      await pickCategory(tester, _fees);
      await chooseIn(tester, _e1, _pf);
      await typeInto(tester, _amount, '40000');
      await settlePreview(tester);

      await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
      await expectLater(tester, meetsGuideline(iOSTapTargetGuideline));
      await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
      await expectLater(tester, meetsGuideline(textContrastGuideline));
      handle.dispose();

      tester.platformDispatcher.textScaleFactorTestValue = 2;
      addTearDown(tester.platformDispatcher.clearTextScaleFactorTestValue);
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
    });
  }
}
