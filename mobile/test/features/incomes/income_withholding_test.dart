import 'dart:async';

import 'package:control_contable/api/generated/models/counterparty_kind.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/api/generated/models/operation_kind.dart';
import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/api/generated/models/warning_code.dart';
import 'package:control_contable/api/generated/models/withholder.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/features/incomes/incomes_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/choice_finders.dart';
import '../../helpers/withholding_fixtures.dart';
import '../taxpayers/pump_onboarding.dart' show inputLabeled, tapText, typeInto;
import '../taxpayers/taxpayers_fakes.dart';
import 'incomes_fakes.dart';
import 'pump_incomes.dart';

const _amount = 'Monto sin IVA';
const _ivaRate = 'Tasa de IVA';
const _q1 = '¿A quién le facturaste?';
const _q2 = '¿Qué le vendiste o qué servicio diste?';
const _q3 = '¿Es una casa o departamento para vivir?';
const _q4 =
    '¿Lo cobraste a través de una plataforma digital (app de transporte, '
    'hospedaje o ventas)?';
const _pm = 'Una empresa (persona moral)';
const _govFederal = 'Gobierno federal (o un organismo federal)';
const _govLocal = 'Gobierno estatal o municipal';
const _pfWithActivity =
    'Una persona con negocio o profesión (persona física con actividad)';
const _pfPublic = 'Una persona sin negocio o público en general';
const _foreign = 'Alguien en el extranjero';
const _rent = 'Renta de un inmueble (casa, local, oficina, terreno)';
const _more = 'Más detalles';
const _cpaWarning =
    'Esta retención aún no la confirma un contador. Úsala como aproximado.';
const _foreignWarning =
    'Las operaciones con el extranjero pueden llevar otros impuestos que la '
    'app todavía no calcula.';
const _nothingWithheld = 'Nadie te retiene impuestos en esta operación.';

final Taxpayer _resico = taxpayer(regimes: [heldRegime('626')]);
final Taxpayer _landlord = taxpayer(regimes: [heldRegime('606')]);
final Taxpayer _moral601 = taxpayer(
  personType: PersonType.moral,
  regimes: [heldRegime('601')],
);
Taxpayer _platforms({bool rfcGiven = true}) => taxpayer(
  regimes: [
    heldRegime(
      '625',
      options: {
        'activities': ['goods_services'],
        'rfc_given_to_platform': rfcGiven,
      },
    ),
  ],
);

void main() {
  late IncomesHarness harness;

  setUp(() => harness = IncomesHarness());

  Future<void> pumpForm(WidgetTester tester, {Taxpayer? taxpayer}) async {
    await harness.pump(tester, taxpayer: taxpayer ?? anaOneRegime);
    unawaited(
      harness.container
          .read(appRouterProvider)
          .push<void>(IncomesRoutes.create),
    );
    await tester.pumpAndSettle();
  }

  /// A user with no income yet in the regime: nothing preselected, and
  /// the regime's defaults ([operation], [viaPlatform]) as the server
  /// answers them.
  void firstTime({
    OperationKind operation = OperationKind.professionalServices,
    bool viaPlatform = false,
  }) => harness.incomes.defaultsAnswer = noWithholdingDefaults.copyWith(
    lastUsed: lastUsed(
      counterpartyKind: null,
      operationKind: operation,
      viaPlatform: viaPlatform,
    ),
  );

  group('WH-11 who you billed (Q1)', () {
    testWidgets('always asked, its six options in plain words, nothing '
        'preselected the first time (OQ-W2)', (tester) async {
      firstTime();
      await pumpForm(tester);

      expect(find.text(_q1), findsOneWidget);
      for (final option in [
        _pm,
        _govFederal,
        _govLocal,
        _pfWithActivity,
        _pfPublic,
        _foreign,
      ]) {
        expect(
          find.descendant(of: groupOf(_q1), matching: find.text(option)),
          findsOneWidget,
        );
      }
      expect(chosenIn(tester, _q1), isNull);
      // The old payer question is gone.
      expect(find.text('¿Quién pagó?'), findsNothing);
    });

    testWidgets('left unanswered it is an error on submit only, and nothing '
        'is sent', (tester) async {
      firstTime();
      await pumpForm(tester);
      await typeInto(tester, _amount, '1000');

      await tapText(tester, 'Guardar');

      expect(find.text('Elige una opción.'), findsOneWidget);
      expect(harness.incomes.creates, isEmpty);
    });

    testWidgets('a returning user starts at the last answers used in the '
        'regime', (tester) async {
      harness.incomes.defaultsAnswer = noWithholdingDefaults.copyWith(
        lastUsed: lastUsed(
          counterpartyKind: CounterpartyKind.pm,
          operationKind: OperationKind.businessServices,
        ),
      );
      await pumpForm(tester);

      expect(chosenIn(tester, _q1), CounterpartyKind.pm.json);
      expect(chosenIn(tester, _q2), OperationKind.businessServices.json);
    });
  });

  group('WH-11 what you sold (Q2) appears only when the payer can '
      'withhold', () {
    for (final (option, shown) in [
      (_pm, true),
      (_govFederal, true),
      (_govLocal, true),
      (_pfWithActivity, false),
      (_pfPublic, false),
      (_foreign, false),
    ]) {
      testWidgets('"$option" ${shown ? 'asks' : 'does not ask'} it', (
        tester,
      ) async {
        firstTime();
        await pumpForm(tester);

        await chooseIn(tester, _q1, option);

        expect(find.text(_q2), shown ? findsOneWidget : findsNothing);
      });
    }

    testWidgets('it starts at the regime default (612 professional '
        'services) and offers no "subcontratación" (WH-07)', (tester) async {
      firstTime();
      await pumpForm(tester);

      await chooseIn(tester, _q1, _pm);

      expect(chosenIn(tester, _q2), OperationKind.professionalServices.json);
      expect(find.textContaining('ubcontrata'), findsNothing);
    });

    testWidgets('the regime default is the server\'s (last_used), never a '
        'table of the app: whatever it answers is preselected', (
      tester,
    ) async {
      // 606 would be rent in the spec's table; the server is the authority.
      firstTime(operation: OperationKind.saleOfGoods);
      await pumpForm(tester, taxpayer: _landlord);

      await chooseIn(tester, _q1, _pm);

      expect(chosenIn(tester, _q2), OperationKind.saleOfGoods.json);
    });

    testWidgets('without the defaults (offline) it starts unanswered and is '
        'omitted, so the server applies its own default', (tester) async {
      harness.incomes.failAlways['defaults'] = const ApiUnreachable();
      await pumpForm(tester, taxpayer: _landlord);
      await chooseIn(tester, _q1, _pm);

      expect(find.text(_q2), findsOneWidget);
      expect(chosenIn(tester, _q2), isNull);
      expect(find.text(_q3), findsNothing);
      await typeInto(tester, _amount, '1000');
      await tapText(tester, 'Guardar');

      expect(harness.incomes.creates.single.operationKind, isNull);
    });
  });

  group('WH-11 residential rent (Q3)', () {
    testWidgets('asked only for the rent of a property', (tester) async {
      firstTime();
      await pumpForm(tester);
      await chooseIn(tester, _q1, _pm);
      expect(find.text(_q3), findsNothing);

      await chooseIn(tester, _q2, _rent);

      expect(find.text(_q3), findsOneWidget);
      expect(find.text('La renta de vivienda no lleva IVA.'), findsOneWidget);
    });

    testWidgets('606 asks it even when what you sold is not asked (rent by '
        'default)', (tester) async {
      firstTime(operation: OperationKind.rentRealEstate);
      await pumpForm(tester, taxpayer: _landlord);

      await chooseIn(tester, _q1, _pfWithActivity);

      expect(find.text(_q2), findsNothing);
      expect(find.text(_q3), findsOneWidget);
    });

    testWidgets('"Sí" makes the income exempt, hides the IVA rate and sends '
        'the fact (LIVA 20 fr. II)', (tester) async {
      firstTime(operation: OperationKind.rentRealEstate);
      await pumpForm(tester, taxpayer: _landlord);
      await chooseIn(tester, _q1, _pm);
      await typeInto(tester, _amount, '30000');

      await chooseIn(tester, _q3, 'Sí');

      expect(find.text(_ivaRate), findsNothing);
      expect(find.text(r'Total con IVA $30,000.00'), findsOneWidget);
      await tapText(tester, 'Guardar');
      final sent = harness.incomes.creates.single;
      expect(sent.residentialRent, isTrue);
      expect(sent.ivaRate, IvaRate.exempt);
      expect(sent.ivaCents, 0);
      expect(sent.operationKind, OperationKind.rentRealEstate);
    });

    testWidgets('back to "No", the IVA rate returns at its rate', (
      tester,
    ) async {
      firstTime(operation: OperationKind.rentRealEstate);
      await pumpForm(tester, taxpayer: _landlord);
      await chooseIn(tester, _q1, _pm);
      await chooseIn(tester, _q3, 'Sí');

      await chooseIn(tester, _q3, 'No');

      expect(find.text(_ivaRate), findsOneWidget);
      expect(chosenIn(tester, _ivaRate), IvaRate.rate16.json);
    });
  });

  group('WH-11 via a platform (Q4)', () {
    testWidgets('625 asks it, "Sí" by default, and then what you sold is '
        'not asked', (tester) async {
      firstTime(viaPlatform: true);
      await pumpForm(tester, taxpayer: _platforms());

      expect(find.text(_q4), findsOneWidget);
      expect(chosenIn(tester, _q4), 'true');
      await chooseIn(tester, _q1, _pm);
      expect(find.text(_q2), findsNothing);

      await chooseIn(tester, _q4, 'No');

      expect(find.text(_q2), findsOneWidget);
    });

    testWidgets('a persona moral finds it under "Más detalles", "No" by '
        'default', (tester) async {
      firstTime();
      await pumpForm(tester, taxpayer: _moral601);
      expect(find.text(_q4), findsNothing);

      await tapText(tester, _more);

      expect(find.text(_q4), findsOneWidget);
      expect(chosenIn(tester, _q4), 'false');
    });

    testWidgets('a persona física outside 625 is never asked', (tester) async {
      firstTime();
      await pumpForm(tester);
      await tapText(tester, _more);

      expect(find.text(_q4), findsNothing);
    });

    testWidgets('625 without the RFC given to the platform warns of the '
        'higher withholding', (tester) async {
      firstTime(viaPlatform: true);
      await pumpForm(tester, taxpayer: _platforms(rfcGiven: false));

      expect(
        find.text(
          'No le diste tu RFC a la plataforma: te retiene 20 % de ISR y todo '
          'el IVA.',
        ),
        findsOneWidget,
      );
    });
  });

  group('WH-11 the derived card', () {
    testWidgets("owner's example: RESICO bills 100,000.00 of professional "
        'services to a company ⇒ ISR 1,250.00, IVA 10,666.67, receives '
        '104,083.33 (US-W1)', (tester) async {
      firstTime();
      harness.incomes.previewAnswer = (_) => ownerExample;
      await pumpForm(tester, taxpayer: _resico);

      await chooseIn(tester, _q1, _pm);
      await typeInto(tester, _amount, '100000');
      await settlePreview(tester);

      final query = harness.incomes.previewRequests.last;
      expect(query.regime, '626');
      expect(query.on, '2026-10-13');
      expect(query.amountCents, 10000000);
      expect(query.ivaRate, IvaRate.rate16);
      expect(query.ivaCents, 1600000);
      expect(query.counterpartyKind, CounterpartyKind.pm);
      expect(query.operationKind, OperationKind.professionalServices);
      expect(query.viaPlatform, isNull);
      expect(
        find.text(r'Te retienen: ISR $1,250.00 · IVA $10,666.67'),
        findsOneWidget,
      );
      expect(find.text(r'Recibes $104,083.33'), findsOneWidget);
      expect(
        find.text(
          'La empresa te retiene 1.25 % de ISR porque estás en RESICO '
          '(LISR art. 113-J).',
        ),
        findsOneWidget,
      );

      await tapText(tester, 'Guardar');
      final sent = harness.incomes.creates.single;
      expect(sent.counterpartyKind, CounterpartyKind.pm);
      expect(sent.operationKind, OperationKind.professionalServices);
      expect(sent.isrWithheldCents, isNull);
      expect(sent.ivaWithheldCents, isNull);
    });

    testWidgets('no withholding field can be typed: the old inputs are '
        'gone and the card is read-only', (tester) async {
      harness.incomes.previewAnswer = (_) => ownerExample;
      await pumpForm(tester, taxpayer: _resico);
      await chooseIn(tester, _q1, _pm);
      await typeInto(tester, _amount, '100000');
      await settlePreview(tester);
      await tapText(tester, _more);

      expect(find.text('ISR retenido (opcional)'), findsNothing);
      expect(find.text('IVA retenido (opcional)'), findsNothing);
      expect(find.text('Retenciones'), findsNothing);
      final card = find.byKey(const ValueKey('withholding-card'));
      expect(
        find.descendant(of: card, matching: find.byType(TextField)),
        findsNothing,
      );
    });

    testWidgets('nobody withholds from the public: it says so', (tester) async {
      await pumpForm(tester);
      await chooseIn(tester, _q1, _pfPublic);
      await typeInto(tester, _amount, '5000');
      await settlePreview(tester);

      expect(find.text(_nothingWithheld), findsOneWidget);
      expect(find.text(r'Recibes $5,800.00'), findsOneWidget);
    });

    testWidgets('before the amount and the payer it says what it needs', (
      tester,
    ) async {
      firstTime();
      await pumpForm(tester);

      expect(
        find.text(
          'Escribe el monto y elige a quién le facturaste para ver si te '
          'retienen impuestos.',
        ),
        findsOneWidget,
      );
      expect(harness.incomes.previewRequests, isEmpty);
    });

    testWidgets('typing is debounced into one request', (tester) async {
      await pumpForm(tester);
      final field = inputLabeled(_amount);

      await tester.enterText(field, '1');
      await tester.pump(const Duration(milliseconds: 100));
      await tester.enterText(field, '10');
      await tester.pump(const Duration(milliseconds: 100));
      await tester.enterText(field, '100');
      await settlePreview(tester);

      expect(harness.incomes.previewRequests, hasLength(1));
      expect(harness.incomes.previewRequests.single.amountCents, 10000);
    });

    testWidgets('no flicker: the last figures stay while the next answer is '
        'in flight', (tester) async {
      firstTime();
      harness.incomes.previewAnswer = (_) => ownerExample;
      await pumpForm(tester, taxpayer: _resico);
      await chooseIn(tester, _q1, _pm);
      await typeInto(tester, _amount, '100000');
      await settlePreview(tester);
      harness.incomes.previewGate = Completer<void>();

      await typeInto(tester, _amount, '100001');
      await settlePreview(tester);

      expect(find.text(r'Recibes $104,083.33'), findsOneWidget);
      harness.incomes.previewGate!.complete();
      await tester.pumpAndSettle();
    });

    testWidgets('offline: the form stays usable and says the server will '
        'derive it on save', (tester) async {
      harness.incomes.failAlways['preview'] = const ApiUnreachable();
      await pumpForm(tester);
      await typeInto(tester, _amount, '1000');
      await settlePreview(tester);

      expect(
        find.text(
          'Sin conexión no podemos calcular la retención. Se calculará al '
          'guardar.',
        ),
        findsOneWidget,
      );
      await tapText(tester, 'Guardar');
      expect(harness.incomes.creates, hasLength(1));
    });

    testWidgets('an unverified rule warns (⚠️ CPA) without blocking', (
      tester,
    ) async {
      harness.incomes.previewAnswer = (_) => preview(
        withholding(
          isrCents: 300000,
          ivaCents: 320000,
          withholder: Withholder.counterparty,
          netCents: 2860000,
          cpaUnverified: true,
        ),
        warnings: const [WarningCode.withholdingCpaUnverified],
      );
      await pumpForm(tester);
      await chooseIn(tester, _q1, _govFederal);
      await typeInto(tester, _amount, '30000');
      await settlePreview(tester);

      expect(find.text(_cpaWarning), findsOneWidget);
    });

    testWidgets('a foreign customer warns at once, even offline', (
      tester,
    ) async {
      harness.incomes.failAlways['preview'] = const ApiUnreachable();
      await pumpForm(tester);

      await chooseIn(tester, _q1, _foreign);

      expect(find.text(_foreignWarning), findsOneWidget);
    });

    testWidgets('the derived amounts are announced politely when they '
        'change', (tester) async {
      firstTime();
      harness.incomes.previewAnswer = (_) => ownerExample;
      await pumpForm(tester, taxpayer: _resico);
      await chooseIn(tester, _q1, _pm);
      tester.takeAnnouncements();

      await typeInto(tester, _amount, '100000');
      await settlePreview(tester);

      final announced = tester.takeAnnouncements();
      expect(
        announced.map((a) => a.message),
        contains(
          r'Te retienen: ISR $1,250.00 · IVA $10,666.67. Recibes $104,083.33',
        ),
      );
      expect(
        announced.map((a) => a.assertiveness),
        everyElement(Assertiveness.polite),
      );
    });

    testWidgets('"¿Tu factura dice otra cantidad?" explains the law wins', (
      tester,
    ) async {
      await pumpForm(tester);

      await tapText(tester, '¿Tu factura dice otra cantidad?');

      expect(
        find.textContaining('La app calcula la retención que marca la ley'),
        findsOneWidget,
      );
    });
  });

  group('WH-11 save and create another', () {
    testWidgets('keeps the answers (context), clears the amounts', (
      tester,
    ) async {
      firstTime();
      await pumpForm(tester);
      await chooseIn(tester, _q1, _pm);
      await chooseIn(tester, _q2, _rent);
      await typeInto(tester, _amount, '1000');

      await tapText(tester, 'Guardar y crear otro');

      expect(chosenIn(tester, _q1), CounterpartyKind.pm.json);
      expect(chosenIn(tester, _q2), OperationKind.rentRealEstate.json);
      await typeInto(tester, _amount, '2000');
      await tapText(tester, 'Guardar');
      final second = harness.incomes.creates.last;
      expect(second.counterpartyKind, CounterpartyKind.pm);
      expect(second.operationKind, OperationKind.rentRealEstate);
    });
  });

  for (final brightness in Brightness.values) {
    testWidgets('the questions and the card meet the tap-target, label and '
        'contrast guidelines, and survive 200 % text (${brightness.name})', (
      tester,
    ) async {
      final handle = tester.ensureSemantics();
      harness.incomes.previewAnswer = (_) => preview(
        withholding(
          isrCents: 125000,
          ivaCents: 1066667,
          withholder: Withholder.counterparty,
          netCents: 10408333,
          cpaUnverified: true,
        ),
        warnings: const [WarningCode.withholdingCpaUnverified],
      );
      await harness.pump(tester, taxpayer: _landlord, brightness: brightness);
      unawaited(
        harness.container
            .read(appRouterProvider)
            .push<void>(IncomesRoutes.create),
      );
      await tester.pumpAndSettle();
      await chooseIn(tester, _q1, _pm);
      await typeInto(tester, _amount, '100000');
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
