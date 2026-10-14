import 'dart:async';

import 'package:control_contable/api/generated/models/counterparty_type.dart';
import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/platform_activity.dart';
import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/api/generated/models/treatment.dart';
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/api/generated/models/warning_code.dart';
import 'package:control_contable/core/access/read_only_access.dart';
import 'package:control_contable/core/forms/form_footer.dart';
import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/features/incomes/incomes_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../taxpayers/pump_onboarding.dart' show inputLabeled, tapText, typeInto;
import '../taxpayers/taxpayers_fakes.dart';
import 'incomes_fakes.dart';
import 'pump_incomes.dart';

const _amount = 'Monto sin IVA';
const _description = 'Descripción (opcional)';
const _collectedOn = 'Fecha de cobro';
const _invoicedOn = 'Fecha de factura';
const _ivaAmount = 'IVA cobrado (opcional)';
const _isrWithheld = 'ISR retenido (opcional)';
const _ivaWithheld = 'IVA retenido (opcional)';
const _uuid = 'Folio fiscal (UUID) (opcional)';
const _readOnlyTitle = 'Tu cuenta es de solo consulta';
const _trialCapTitle = 'Llegaste al límite de ingresos de prueba';

String _textOf(WidgetTester tester, String label) =>
    tester.widget<TextField>(inputLabeled(label)).controller!.text;

Future<void> _save(WidgetTester tester) => tapText(tester, 'Guardar');

Future<void> _saveAndAnother(WidgetTester tester) =>
    tapText(tester, 'Guardar y crear otro');

void main() {
  late IncomesHarness harness;

  setUp(() => harness = IncomesHarness());

  Future<void> pumpForm(
    WidgetTester tester, {
    bool enabled = true,
    Brightness brightness = Brightness.light,
    Taxpayer? taxpayer,
  }) async {
    // Opened from the list, as the app does (the list is where it returns).
    await harness.pump(
      tester,
      enabled: enabled,
      brightness: brightness,
      taxpayer: taxpayer ?? anaOneRegime,
      location: enabled ? IncomesRoutes.list : IncomesRoutes.create,
    );
    if (!enabled) return;
    unawaited(
      harness.container
          .read(appRouterProvider)
          .push<void>(IncomesRoutes.create),
    );
    await tester.pumpAndSettle();
  }

  /// US-I1: 10,000.00 from a persona moral.
  Future<void> fillUsI1(WidgetTester tester) async {
    await typeInto(tester, _description, 'Diseño de logotipo');
    await typeInto(tester, _amount, '10000');
    await tapText(tester, 'Persona moral');
  }

  group('M3-09 layout and defaults', () {
    testWidgets('a page with the list columns first (date, description, '
        'amount), defaults prefilled, the rest collapsed', (tester) async {
      await pumpForm(tester);

      expect(find.text('Nuevo ingreso'), findsOneWidget);
      expect(_textOf(tester, _collectedOn), '13/10/2026');
      final order = [
        _collectedOn,
        _description,
        _amount,
        'Tasa de IVA',
        '¿Quién pagó?',
      ];
      for (var i = 1; i < order.length; i++) {
        expect(
          tester.getTopLeft(find.text(order[i - 1]).first).dy,
          lessThan(tester.getTopLeft(find.text(order[i]).first).dy),
          reason: '${order[i - 1]} before ${order[i]}',
        );
      }
      // One regime: preselected, not asked.
      expect(find.text('Régimen'), findsNothing);
      // Progressive disclosure: collapsed until opened.
      expect(find.text(_uuid), findsNothing);
      expect(find.text(_isrWithheld), findsNothing);
      expect(find.text('Más detalles'), findsOneWidget);
      // Defaults: IVA 16 %, persona física the first time.
      final request = harness.incomes.defaultsRequests.last;
      expect(request.regime, '612');
      expect(request.counterparty, CounterpartyType.personaFisica);
      expect(request.on, '2026-10-13');
      final rate16 = tester.widget<RadioListTile<String>>(
        find.widgetWithText(RadioListTile<String>, '16 %'),
      );
      expect(rate16.value, IvaRate.rate16.json);
    });

    testWidgets('the total updates live as the amount is typed (integer '
        'centavos)', (tester) async {
      await pumpForm(tester);

      await typeInto(tester, _amount, '10,000.00');

      expect(find.text(r'Total con IVA $11,600.00'), findsOneWidget);
    });

    testWidgets('a persona moral payer opens the withholdings prefilled '
        'from the defaults and the net follows (US-I1)', (tester) async {
      await pumpForm(tester);

      await fillUsI1(tester);

      expect(
        harness.incomes.defaultsRequests.last.counterparty,
        CounterpartyType.personaMoral,
      );
      expect(_textOf(tester, _isrWithheld), '1,000.00');
      expect(_textOf(tester, _ivaWithheld), '1,066.67');
      expect(
        find.text(r'IVA $1,600.00 · Retenciones −$2,066.67 · Neto $9,533.33'),
        findsOneWidget,
      );
    });

    testWidgets('only the IVA rates allowed on the date are offered', (
      tester,
    ) async {
      harness.incomes.defaultsBy[CounterpartyType.personaFisica] =
          noWithholdingDefaults.copyWith(
            ivaRatesAllowed: [IvaRate.rate16, IvaRate.rate0, IvaRate.exempt],
          );
      await pumpForm(tester);

      expect(find.text('8 % (región fronteriza)'), findsNothing);
      expect(find.text('Exento'), findsOneWidget);
    });

    testWidgets('an exempt income carries no IVA', (tester) async {
      await pumpForm(tester);
      await typeInto(tester, _amount, '500');

      await tapText(tester, 'Exento');

      expect(find.text(r'Total con IVA $500.00'), findsOneWidget);
    });

    testWidgets('with two regimes the regime is asked, preselected, and '
        'the defaults follow it', (tester) async {
      await pumpForm(tester, taxpayer: anaTwoRegimes);

      expect(find.text('Régimen'), findsOneWidget);
      await tapText(tester, 'Régimen 606');

      expect(harness.incomes.defaultsRequests.last.regime, '606');
    });

    testWidgets('an accrual regime (601) books by the invoice date', (
      tester,
    ) async {
      await pumpForm(
        tester,
        taxpayer: taxpayer(
          personType: PersonType.moral,
          regimes: [heldRegime('601')],
        ),
      );
      expect(find.text(_invoicedOn), findsOneWidget);
      expect(find.text(_collectedOn), findsNothing);

      await typeInto(tester, _amount, '1000');
      await _save(tester);

      final sent = harness.incomes.creates.single;
      expect(sent.invoicedOn, '2026-10-13');
      expect(sent.collectedOn, isNull);
    });

    // M5-30 (M5-29 QA, PM-601-IVA-01): a 601 income counts for IVA only
    // once collected (LIVA 1-B); its collection date lives in "Más
    // detalles" and must reach the request.
    testWidgets('an accrual regime (601) sends the collection date given '
        'under "Más detalles"', (tester) async {
      await pumpForm(
        tester,
        taxpayer: taxpayer(
          personType: PersonType.moral,
          regimes: [heldRegime('601')],
        ),
      );
      await typeInto(tester, _invoicedOn, '05102026');
      await typeInto(tester, _amount, '1000');
      await tapText(tester, 'Más detalles');
      await typeInto(tester, '$_collectedOn (opcional)', '12102026');
      await _save(tester);

      final sent = harness.incomes.creates.single;
      expect(sent.invoicedOn, '2026-10-05');
      expect(sent.collectedOn, '2026-10-12');
    });
  });

  group('M3-09 save', () {
    testWidgets('Guardar sends integer centavos with the client id, goes '
        'back to the list and confirms', (tester) async {
      await pumpForm(tester);
      await fillUsI1(tester);

      await _save(tester);

      final sent = harness.incomes.creates.single;
      expect(sent.id, harness.clientIds.first);
      expect(sent.regimeCode, '612');
      // ignore: deprecated_member_use_from_same_package, deprecated until WH-11 (ADR-0008).
      expect(sent.counterpartyType, CounterpartyType.personaMoral);
      expect(sent.amountCents, 1000000);
      expect(sent.ivaRate, IvaRate.rate16);
      expect(sent.ivaCents, 160000);
      expect(sent.isrWithheldCents, 100000);
      expect(sent.ivaWithheldCents, 106667);
      expect(sent.collectedOn, '2026-10-13');
      expect(sent.description, 'Diseño de logotipo');
      expect(harness.location, IncomesRoutes.list);
      expect(find.text('Ingreso guardado'), findsOneWidget);
      expect(find.text('Diseño de logotipo'), findsOneWidget);
    });

    testWidgets('a second tap while saving is ignored: the primary stays '
        'enabled and busy, one request', (tester) async {
      harness.incomes.createGate = Completer<void>();
      await pumpForm(tester);
      await typeInto(tester, _amount, '1000');

      await tester.tap(find.widgetWithText(FilledButton, 'Guardar'));
      await tester.pump();
      final busy = find.ancestor(
        of: find.bySemanticsLabel('Guardando…'),
        matching: find.byType(FilledButton),
      );
      expect(tester.widget<FilledButton>(busy).onPressed, isNotNull);
      await tester.tap(busy);
      await tester.pump();

      harness.incomes.createGate!.complete();
      await tester.pumpAndSettle();
      expect(harness.incomes.creates, hasLength(1));
    });

    testWidgets('after a lost answer the retry resends the same id, so it '
        'never duplicates', (tester) async {
      harness.incomes.failNext['create'] = const ApiUnreachable();
      await pumpForm(tester);
      await typeInto(tester, _amount, '1000');

      await _save(tester);
      expect(
        find.text(
          'No se pudo guardar por falta de conexión. Vuelve a tocar '
          'Guardar: no se duplicará.',
        ),
        findsOneWidget,
      );
      expect(parseCentavos(_textOf(tester, _amount)), 100000);

      await _save(tester);

      expect(harness.incomes.creates, hasLength(2));
      expect(harness.incomes.creates.map((c) => c.id).toSet(), {
        harness.clientIds.first,
      });
      expect(harness.incomes.incomes, hasLength(1));
    });

    testWidgets('"Guardar y crear otro" keeps the context (dates, regime, '
        'who paid, rate), clears the rest, refocuses the first cleared '
        'field and starts a new id', (tester) async {
      await pumpForm(tester);
      await typeInto(tester, _collectedOn, '05102026');
      await fillUsI1(tester);

      await _saveAndAnother(tester);

      expect(harness.location, IncomesRoutes.create);
      expect(find.text('Ingreso guardado'), findsOneWidget);
      expect(_textOf(tester, _collectedOn), '05/10/2026');
      expect(_textOf(tester, _description), isEmpty);
      expect(_textOf(tester, _amount), isEmpty);
      final description = tester.widget<TextField>(inputLabeled(_description));
      expect(description.focusNode!.hasFocus, isTrue);

      await typeInto(tester, _amount, '2000');
      await _save(tester);

      final [first, second] = harness.incomes.creates;
      expect(first.collectedOn, '2026-10-05');
      expect(second.collectedOn, '2026-10-05');
      // ignore: deprecated_member_use_from_same_package, deprecated until WH-11 (ADR-0008).
      expect(second.counterpartyType, CounterpartyType.personaMoral);
      expect(second.isrWithheldCents, 20000);
      expect(second.description, isNull);
      expect(first.id, isNot(second.id));
    });

    testWidgets('required fields left empty show on submit only, and '
        'nothing is sent', (tester) async {
      await pumpForm(tester);

      expect(find.text('Escribe este dato.'), findsNothing);
      await _save(tester);

      expect(find.text('Revisa 1 dato antes de guardar'), findsOneWidget);
      expect(find.text('Escribe este dato.'), findsOneWidget);
      expect(harness.incomes.creates, isEmpty);
    });

    testWidgets('the folio fiscal is checked on blur and sent upper-case', (
      tester,
    ) async {
      await pumpForm(tester);
      await typeInto(tester, _amount, '1000');
      await tapText(tester, 'Más detalles');

      await typeInto(tester, _uuid, '123');
      // Leaving the field (focus moves on) checks it.
      await tester.ensureVisible(inputLabeled(_description));
      await tester.tap(inputLabeled(_description));
      await tester.pumpAndSettle();
      expect(
        find.textContaining('Escribe el folio como aparece en el CFDI'),
        findsOneWidget,
      );

      await typeInto(tester, _uuid, 'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d');
      await _save(tester);
      expect(
        harness.incomes.creates.single.cfdiUuid,
        'A1B2C3D4-E5F6-4A7B-8C9D-0E1F2A3B4C5D',
      );
    });

    testWidgets('a server verdict lands on its field, opening its section '
        '(iva_amount_mismatch)', (tester) async {
      harness.incomes.failNext['create'] = ApiRejected(
        code: ErrorCode.businessRuleViolation,
        details: [detail('iva_cents', 'iva_amount_mismatch')],
      );
      await pumpForm(tester);
      await typeInto(tester, _amount, '1000');

      await _save(tester);

      expect(find.text(_ivaAmount), findsOneWidget);
      expect(
        find.text(
          'El IVA no corresponde a la tasa elegida. Revísalo en el CFDI.',
        ),
        findsWidgets,
      );
      expect(harness.location, IncomesRoutes.create);
    });

    testWidgets('warnings never block: saved, and the user reads them', (
      tester,
    ) async {
      harness.incomes.warnNext = const [
        Warning(
          code: WarningCode.duplicateCfdiUuid,
          field: 'cfdi_uuid',
          message: 'Duplicate CFDI UUID.',
        ),
      ];
      await pumpForm(tester);
      await typeInto(tester, _amount, '1000');

      await _saveAndAnother(tester);

      expect(harness.incomes.incomes, hasLength(1));
      expect(
        find.textContaining('Ya hay otro ingreso con este folio fiscal'),
        findsOneWidget,
      );
    });
  });

  // QA M3/M4 D2: after *Guardar y crear otro* the warning toast stays until
  // closed, on the form itself; it must not cover the footer's actions.
  testWidgets('a persistent warning toast leaves the form footer clear', (
    tester,
  ) async {
    harness.incomes.warnNext = const [
      Warning(
        code: WarningCode.duplicateCfdiUuid,
        field: 'cfdi_uuid',
        message: 'Duplicate CFDI UUID.',
      ),
    ];
    await pumpForm(tester);
    await typeInto(tester, _amount, '1000');

    await _saveAndAnother(tester);

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

  group('M3-09 edge cases', () {
    testWidgets('defaults that fail to load leave the form usable', (
      tester,
    ) async {
      harness.incomes.failAlways['defaults'] = const ApiUnreachable();
      await pumpForm(tester);

      expect(
        find.textContaining('No se pudieron cargar las tasas sugeridas'),
        findsOneWidget,
      );
      await typeInto(tester, _amount, '1000');
      await _save(tester);
      expect(harness.incomes.creates.single.ivaCents, 16000);
    });

    testWidgets('a reused id (409) says the income was already saved', (
      tester,
    ) async {
      harness.incomes.failNext['create'] = const ApiRejected(
        code: ErrorCode.conflict,
      );
      await pumpForm(tester);
      await typeInto(tester, _amount, '1000');

      await _save(tester);

      expect(
        find.text(
          'Este ingreso ya se había guardado con otros datos. Revisa la '
          'lista de ingresos.',
        ),
        findsOneWidget,
      );
    });

    testWidgets('a verdict on no visible field becomes a form message', (
      tester,
    ) async {
      harness.incomes.failNext['create'] = ApiRejected(
        code: ErrorCode.businessRuleViolation,
        details: [detail('taxpayerId', 'taxpayer_archived')],
      );
      await pumpForm(tester);
      await typeInto(tester, _amount, '1000');

      await _save(tester);

      expect(find.textContaining('archivado'), findsOneWidget);
    });

    testWidgets('625 with several activities asks which one, and sends it', (
      tester,
    ) async {
      await pumpForm(
        tester,
        taxpayer: taxpayer(
          regimes: [
            heldRegime(
              '625',
              options: {
                'activities': ['transport_delivery', 'lodging'],
              },
            ),
          ],
        ),
      );
      await typeInto(tester, _amount, '15000');

      await _save(tester);
      expect(find.text('Elige una opción.'), findsOneWidget);
      expect(harness.incomes.creates, isEmpty);

      await tapText(tester, 'Hospedaje');
      await _save(tester);
      expect(
        harness.incomes.creates.single.platformActivity,
        PlatformActivity.lodging,
      );
      expect(
        harness.incomes.defaultsRequests.last.activity,
        PlatformActivity.lodging,
      );
    });

    testWidgets('an informational regime says the income does not change '
        'the estimate (OQ-4)', (tester) async {
      await pumpForm(
        tester,
        taxpayer: taxpayer(
          regimes: [heldRegime('605', treatment: Treatment.informational)],
        ),
      );

      expect(
        find.text(
          'Este régimen es informativo: el ingreso no cambia tu estimación.',
        ),
        findsOneWidget,
      );
    });
  });

  group('M3-09 access', () {
    testWidgets('a spent trial cap (402 incomes_remaining) names the '
        'incomes cap and keeps the form', (tester) async {
      harness.incomes.failNext['create'] = ApiEntitlementRequired(
        details: [detail('incomes_remaining', 'trial_cap_reached')],
      );
      await pumpForm(tester);
      await typeInto(tester, _amount, '1000');

      await _save(tester);

      expect(find.text(_trialCapTitle), findsOneWidget);
      expect(find.textContaining('La prueba incluye 5 ingresos'), findsOne);
      expect(harness.container.read(readOnlyAccessProvider), isFalse);
      await tapText(tester, 'Entendido');
      expect(parseCentavos(_textOf(tester, _amount)), 100000);
    });

    testWidgets('a read-only account (402) gets the read-only message', (
      tester,
    ) async {
      harness.incomes.failNext['create'] = const ApiEntitlementRequired();
      await pumpForm(tester);
      await typeInto(tester, _amount, '1000');

      await _save(tester);

      expect(find.text(_readOnlyTitle), findsOneWidget);
      expect(harness.container.read(readOnlyAccessProvider), isTrue);
    });

    testWidgets('module_incomes off: the form is not found', (tester) async {
      await pumpForm(tester, enabled: false);

      expect(harness.location, AppRoutes.notFound);
      expect(harness.incomes.calls, isEmpty);
    });
  });

  group('M3-09 leaving', () {
    testWidgets('Cancelar with nothing typed leaves at once', (tester) async {
      await pumpForm(tester);

      await tapText(tester, 'Cancelar');

      expect(harness.location, isNot(IncomesRoutes.create));
    });

    testWidgets('Cancelar with changes asks first', (tester) async {
      await pumpForm(tester);
      await typeInto(tester, _amount, '1000');

      await tapText(tester, 'Cancelar');
      expect(find.text('¿Descartar este ingreso?'), findsOneWidget);
      await tapText(tester, 'Seguir editando');
      expect(harness.location, IncomesRoutes.create);

      await tapText(tester, 'Cancelar');
      await tapText(tester, 'Descartar');
      expect(harness.location, isNot(IncomesRoutes.create));
    });

    // QA M3/M4 D-8: the kept context fields of the next record are not
    // unsaved changes; what is typed after them is.
    testWidgets('after "Guardar y crear otro", Cancelar on the untouched '
        'next record leaves at once; a typed value still asks', (tester) async {
      await pumpForm(tester);
      await typeInto(tester, _collectedOn, '05102026');
      await fillUsI1(tester);
      await _saveAndAnother(tester);

      await typeInto(tester, _amount, '1000');
      await tapText(tester, 'Cancelar');
      expect(find.text('¿Descartar este ingreso?'), findsOneWidget);
      await tapText(tester, 'Seguir editando');

      await typeInto(tester, _amount, '');
      await tapText(tester, 'Cancelar');
      expect(find.text('¿Descartar este ingreso?'), findsNothing);
      expect(harness.location, isNot(IncomesRoutes.create));
    });
  });

  for (final brightness in Brightness.values) {
    testWidgets('meets the tap-target, label and contrast guidelines, and '
        'survives 200 % text (${brightness.name})', (tester) async {
      final handle = tester.ensureSemantics();
      await pumpForm(tester, brightness: brightness, taxpayer: anaTwoRegimes);
      await tapText(tester, 'Persona moral');
      await tapText(tester, 'Más detalles');

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
