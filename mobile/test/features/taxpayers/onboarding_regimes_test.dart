import 'dart:async';

import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/regime_option_def_type.dart';
import 'package:control_contable/api/generated/models/tax_input_def_kind.dart';
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/taxpayers/presentation/onboarding_screen.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'pump_onboarding.dart';
import 'taxpayers_fakes.dart';

const _next = 'Siguiente';
const _create = 'Crear contribuyente';
const _legend =
    'Esto es un aproximado. '
    'No todos los gastos aplican para tu giro comercial.';
const _incompatible =
    'No se puede combinar con el 612: '
    'El RESICO sustituye a este régimen.';

Finder _regimeRow(String code) => find
    .ancestor(
      of: find.textContaining('$code ·'),
      matching: find.byType(InkWell),
    )
    .first;

Future<void> _pick(WidgetTester tester, String code) async {
  await tester.ensureVisible(_regimeRow(code));
  await tester.pumpAndSettle();
  await tester.tap(_regimeRow(code));
  await tester.pumpAndSettle();
}

void main() {
  late OnboardingHarness harness;

  setUp(() => harness = OnboardingHarness());

  Future<void> openRegimes(
    WidgetTester tester, {
    String card = 'Persona física',
    String rfc = 'GOMA800101AB6',
    Set<AppModule> alsoEnabled = const {},
    Brightness brightness = Brightness.light,
  }) async {
    await harness.pump(
      tester,
      alsoEnabled: alsoEnabled,
      brightness: brightness,
    );
    await passIdentity(tester, card: card, rfc: rfc);
  }

  group('step 3: regimes by person type', () {
    testWidgets('only the chosen person type catalog, grouped by support '
        'level with plain explanations and the legend', (tester) async {
      await openRegimes(tester);

      expect(harness.repository.catalogCalls, [PersonType.fisica]);
      expect(find.text('Cálculo completo'), findsOneWidget);
      expect(find.text('Cálculo parcial'), findsOneWidget);
      expect(find.text('Solo explicación'), findsOneWidget);
      expect(find.textContaining('estima cada mes el ISR'), findsOneWidget);
      expect(find.textContaining('no calcula impuestos'), findsOneWidget);
      expect(find.text(_legend), findsOneWidget);
      expect(
        find.text('Tu patrón calcula y retiene tu ISR.'),
        findsOneWidget,
        reason: 'an informational regime explains itself before selection',
      );
      expect(find.textContaining('601 ·'), findsNothing);
    });

    testWidgets('a persona moral sees the moral catalog, single-select', (
      tester,
    ) async {
      await openRegimes(tester, card: 'Persona moral', rfc: 'ABC680524P76');

      expect(harness.repository.catalogCalls, [PersonType.moral]);
      await _pick(tester, '601');
      await _pick(tester, '626');

      final selected = tester.widget<Radio<String>>(
        find.descendant(
          of: _regimeRow('626'),
          matching: find.byType(Radio<String>),
        ),
      );
      expect(selected.value, '626');
      final group = tester.widget<RadioGroup<String>>(
        find.byType(RadioGroup<String>),
      );
      expect(group.groupValue, '626', reason: 'picking one drops the other');
    });

    testWidgets('a failed load offers a retry', (tester) async {
      harness.repository.catalogFailure = const ApiUnreachable();
      await openRegimes(tester);

      expect(find.text('No se pudieron cargar los regímenes.'), findsOneWidget);
      await tapText(tester, 'Reintentar');

      expect(find.text('Cálculo completo'), findsOneWidget);
      expect(harness.repository.catalogCalls, hasLength(2));
    });

    testWidgets('an empty catalog says so', (tester) async {
      harness.repository.catalogs[PersonType.fisica] = const [];
      await openRegimes(tester);

      expect(
        find.text('No hay regímenes disponibles para este tipo de persona.'),
        findsOneWidget,
      );
    });

    testWidgets('the catalog shows a status while it loads', (tester) async {
      await harness.pump(tester);
      await tapText(tester, 'Persona física');
      await typeInto(tester, 'Nombre para identificarlo', 'Ana');
      await typeInto(tester, 'RFC', 'GOMA800101AB6');
      await tester.tap(find.text(_next));
      await tester.pump();

      expect(find.text('Cargando regímenes…'), findsOneWidget);
      await tester.pumpAndSettle();
    });

    testWidgets('nothing chosen: the error shows on submit', (tester) async {
      await openRegimes(tester);

      await tapText(tester, _next);

      expect(find.text('Elige al menos un régimen.'), findsWidgets);
      expect(find.text('Paso 3 de 4'), findsOneWidget);
    });

    testWidgets('an incompatible pair shows its reason inline and blocks '
        'Siguiente', (tester) async {
      await openRegimes(tester);
      await _pick(tester, '612');
      await _pick(tester, '626');

      expect(find.text(_incompatible), findsOneWidget);

      await tapText(tester, _next);
      expect(find.text('Paso 3 de 4'), findsOneWidget);
      expect(find.textContaining('no se pueden combinar'), findsWidgets);

      await _pick(tester, '612');
      expect(find.text(_incompatible), findsNothing);
    });

    testWidgets('an incompatible pick is announced once, when it appears', (
      tester,
    ) async {
      await openRegimes(tester);
      await _pick(tester, '612');
      tester.takeAnnouncements();

      await _pick(tester, '626');
      expect(tester.takeAnnouncements().map((a) => a.message), [_incompatible]);

      await _pick(tester, '606');
      expect(
        tester.takeAnnouncements(),
        isEmpty,
        reason: 'a conflict already told is not repeated',
      );
    });

    testWidgets('a regime without options creates right away', (tester) async {
      await openRegimes(tester);
      await _pick(tester, '626');

      await tapText(tester, _create);

      expect(
        harness.repository.creates.single.regimes.single.regimeCode,
        '626',
      );
      expect(find.byType(OnboardingScreen), findsNothing);
    });
  });

  group('step 4: per-regime options', () {
    testWidgets('606 is prefilled blind_35; 625 asks for activities with none '
        'preselected and shows its boolean defaults', (tester) async {
      await openRegimes(tester);
      await _pick(tester, '606');
      await _pick(tester, '625');
      await tapText(tester, _next);

      expect(find.text('Paso 4 de 4'), findsOneWidget);
      expect(find.text('Régimen 1 de 2: 606 · Arrendamiento'), findsOneWidget);
      final mode = tester.widget<RadioGroup<String>>(
        find.byType(RadioGroup<String>),
      );
      expect(mode.groupValue, 'blind_35');
      expect(
        find.text('Deducción ciega: 35 % de los ingresos'),
        findsOneWidget,
      );
      await tapText(tester, _next);

      expect(
        find.text('Régimen 2 de 2: 625 · Plataformas tecnológicas'),
        findsOneWidget,
      );
      final boxes = tester.widgetList<CheckboxListTile>(
        find.byType(CheckboxListTile),
      );
      expect(boxes.map((b) => b.value), [false, false, false]);
      expect(_switchValue(tester, 'Pago definitivo'), isFalse);
      expect(_switchValue(tester, 'Di mi RFC a la plataforma'), isTrue);

      await tapText(tester, _create);
      expect(find.text('Elige una opción.'), findsWidgets);
      expect(harness.repository.creates, isEmpty);

      await tapText(tester, 'Hospedaje');
      await tapText(tester, _create);

      final request = harness.repository.creates.single;
      expect(request.id, testClientId);
      expect(request.regimes.first.options, {'deduction_mode': 'blind_35'});
      expect(request.regimes.last.options, {
        'activities': ['lodging'],
        'definitive_option': false,
        'rfc_given_to_platform': true,
      });
      expect(find.byType(OnboardingScreen), findsNothing);
    });

    testWidgets('back from options keeps the regimes and the options', (
      tester,
    ) async {
      await openRegimes(tester);
      await _pick(tester, '606');
      await tapText(tester, _next);
      await tapText(tester, 'Gastos reales con factura');

      await tester.tap(find.byTooltip('Atrás'));
      await tester.pumpAndSettle();
      final row = tester.widget<Checkbox>(
        find.descendant(of: _regimeRow('606'), matching: find.byType(Checkbox)),
      );
      expect(row.value, isTrue);

      await tapText(tester, _next);
      final mode = tester.widget<RadioGroup<String>>(
        find.byType(RadioGroup<String>),
      );
      expect(mode.groupValue, 'actual');
    });

    testWidgets('a rate option takes es-MX decimals and checks its bounds on '
        'blur', (tester) async {
      harness.repository.catalogs[PersonType.fisica] = [
        regime(
          '612',
          name: 'Actividades',
          optionDefs: [
            optionDef(
              'withholding',
              label: 'Tasa de retención',
              type: RegimeOptionDefType.rate,
              min: '0.1',
              max: '0.5',
            ),
          ],
        ),
      ];
      await openRegimes(tester);
      await _pick(tester, '612');
      await tapText(tester, _next);
      expect(
        find.text('Un número entre 0.1 y 0.5, con punto decimal.'),
        findsOneWidget,
      );

      await typeInto(tester, 'Tasa de retención', '0,9');
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();
      expect(find.text('Escribe un número entre 0.1 y 0.5.'), findsWidgets);

      await typeInto(tester, 'Tasa de retención', 'abc');
      expect(
        find.text('Escribe un número con punto decimal, por ejemplo 0.25.'),
        findsWidgets,
      );

      await typeInto(tester, 'Tasa de retención', ',25');
      await tapText(tester, _create);

      expect(harness.repository.creates.single.regimes.single.options, {
        'withholding': '0.25',
      });
    });

    testWidgets('a server verdict on an option lands on its field', (
      tester,
    ) async {
      harness.repository.createAnswers.add(
        ApiValidationFailure([
          detail('regimes.0.options.rif_year', 'invalid_option'),
        ]),
      );
      await openRegimes(tester);
      await _pick(tester, '621');
      await tapText(tester, _next);
      await tapText(tester, 'Año 2');
      await tapText(tester, _create);

      expect(find.text('Revisa esta opción.'), findsWidgets);
      await tapText(tester, 'Año 3');
      expect(find.text('Revisa esta opción.'), findsNothing);
    });
  });

  group('create outcomes', () {
    testWidgets('402: a read-only user sees the paywall message, no create '
        'path', (tester) async {
      harness.repository.createAnswers.add(const ApiEntitlementRequired());
      await openRegimes(tester);
      await _pick(tester, '626');
      await tapText(tester, _create);

      expect(find.text('Tu cuenta es de solo consulta'), findsOneWidget);
      expect(find.textContaining('requiere un plan'), findsOneWidget);
      expect(find.text(_create), findsNothing);

      await tapText(tester, 'Volver al inicio');
      expect(find.byType(OnboardingScreen), findsNothing);
    });

    testWidgets('after a 402, opening the onboarding again shows the paywall '
        'at once: no step to fill, no second create', (tester) async {
      harness.repository.createAnswers.add(const ApiEntitlementRequired());
      await openRegimes(tester);
      await _pick(tester, '626');
      await tapText(tester, _create);
      await tapText(tester, 'Volver al inicio');

      await tapText(tester, 'Registrar contribuyente');

      expect(find.text('Tu cuenta es de solo consulta'), findsOneWidget);
      expect(find.text('Persona física'), findsNothing);
      expect(find.textContaining('Paso 1'), findsNothing);
      expect(harness.repository.creates, hasLength(1));
    });

    testWidgets('402: the paywall is announced to screen readers, not only '
        'drawn (a live region alone stays silent)', (tester) async {
      harness.repository.createAnswers.add(const ApiEntitlementRequired());
      await openRegimes(tester);
      await _pick(tester, '626');
      tester.takeAnnouncements();

      await tapText(tester, _create);

      final announced = tester.takeAnnouncements().map((a) => a.message);
      expect(announced, contains(startsWith('Tu cuenta es de solo consulta')));
    });

    testWidgets('back from regimes keeps a pick not yet submitted', (
      tester,
    ) async {
      await openRegimes(tester);
      await _pick(tester, '626');

      await tester.tap(find.byTooltip('Atrás'));
      await tester.pumpAndSettle();
      await tapText(tester, _next);

      final row = tester.widget<Checkbox>(
        find.descendant(of: _regimeRow('626'), matching: find.byType(Checkbox)),
      );
      expect(row.value, isTrue);
    });

    testWidgets('409 duplicate RFC: back on step 2 with the field error', (
      tester,
    ) async {
      harness.repository.createAnswers.add(
        const ApiRejected(code: ErrorCode.rfcAlreadyRegistered),
      );
      await openRegimes(tester);
      await _pick(tester, '626');
      await tapText(tester, _create);

      expect(find.text('Paso 2 de 4'), findsOneWidget);
      expect(
        find.textContaining('Ya registraste un contribuyente con este RFC'),
        findsWidgets,
      );

      await typeInto(tester, 'RFC', 'GOMA800101AB7');
      expect(find.textContaining('Ya registraste'), findsNothing);
    });

    testWidgets('the busy create ignores a second tap', (tester) async {
      final gate = harness.repository.createGate = Completer<void>();
      await openRegimes(tester);
      await _pick(tester, '626');
      await tester.ensureVisible(find.text(_create));
      await tester.tap(find.text(_create));
      await tester.pump();

      await tester.tap(find.text(_create));
      await tester.pump();
      expect(harness.repository.creates, hasLength(1));

      gate.complete();
      await tester.pumpAndSettle();
    });

    for (final (failure, message) in [
      (const ApiUnreachable(), 'No hay conexión.'),
      (const ApiRateLimited(null), 'Hubo demasiados cambios seguidos.'),
      (
        const ApiRejected(code: ErrorCode.taxpayerLimitReached),
        'Ya tienes 10 contribuyentes',
      ),
      (
        const ApiRejected(code: ErrorCode.invalidRegimeCombination),
        'Esta combinación de regímenes no se puede registrar.',
      ),
      (
        ApiValidationFailure([detail('regimes.0.regime_code', 'x')]),
        'Uno de los regímenes elegidos no se aceptó.',
      ),
      (const ApiUnexpected(), 'Algo salió mal.'),
    ]) {
      testWidgets('${failure.runtimeType} shows a form notice', (tester) async {
        harness.repository.createAnswers.add(failure);
        await openRegimes(tester);
        await _pick(tester, '626');
        await tapText(tester, _create);

        expect(find.textContaining(message), findsOneWidget);
        expect(find.byType(OnboardingScreen), findsOneWidget);
      });
    }
  });

  group('after create: non-blocking warnings', () {
    const mismatch = Warning(
      code: 'rfc_check_digit_mismatch',
      field: 'rfc',
      message: 'Check digit mismatch.',
    );
    const cpa = Warning(
      code: 'regime_combination_needs_cpa_review',
      field: 'regimes.1.regime_code',
      message: 'Combinación por confirmar con un contador.',
    );

    testWidgets('a check-digit mismatch asks to confirm and continue', (
      tester,
    ) async {
      harness.repository.createAnswers.add(taxpayer(warnings: [mismatch]));
      await openRegimes(tester);
      await _pick(tester, '626');
      await tapText(tester, _create);

      expect(find.text('Revisa que tu RFC esté bien escrito'), findsOneWidget);
      expect(find.textContaining('GOMA******AB6'), findsOneWidget);
      expect(find.text('GOMA800101AB6'), findsOneWidget, reason: 'editable');

      await tapText(tester, 'Continuar');

      expect(harness.repository.corrections, isEmpty);
      expect(find.byType(OnboardingScreen), findsNothing);
    });

    testWidgets('a corrected RFC is sent before continuing', (tester) async {
      harness.repository.createAnswers.add(taxpayer(warnings: [mismatch]));
      await openRegimes(tester);
      await _pick(tester, '626');
      await tapText(tester, _create);

      await typeInto(tester, 'RFC', 'goma800101ab7');
      await tapText(tester, 'Continuar');

      expect(harness.repository.corrections.single.rfc, 'GOMA800101AB7');
      expect(find.byType(OnboardingScreen), findsNothing);
    });

    testWidgets('a refused correction shows on the RFC field', (tester) async {
      harness.repository
        ..createAnswers.add(taxpayer(warnings: [mismatch]))
        ..correctAnswers.add(
          const ApiRejected(code: ErrorCode.rfcAlreadyRegistered),
        );
      await openRegimes(tester);
      await _pick(tester, '626');
      await tapText(tester, _create);

      await typeInto(tester, 'RFC', 'GOMA800101AB7');
      await tapText(tester, 'Continuar');

      expect(find.textContaining('Ya registraste'), findsWidgets);
      expect(find.byType(OnboardingScreen), findsOneWidget);
    });

    testWidgets('a correction that cannot be sent shows a notice', (
      tester,
    ) async {
      harness.repository
        ..createAnswers.add(taxpayer(warnings: [mismatch]))
        ..correctAnswers.add(const ApiUnreachable());
      await openRegimes(tester);
      await _pick(tester, '626');
      await tapText(tester, _create);

      await typeInto(tester, 'RFC', 'GOMA800101AB7');
      await tapText(tester, 'Continuar');

      expect(find.textContaining('No hay conexión.'), findsOneWidget);
    });

    testWidgets('a CPA-review pair is a warning with the server reason', (
      tester,
    ) async {
      harness.repository.createAnswers.add(taxpayer(warnings: [cpa]));
      await openRegimes(tester);
      await _pick(tester, '626');
      await tapText(tester, _create);

      expect(
        find.text('Combinación pendiente de revisión por un contador'),
        findsOneWidget,
      );
      expect(find.text(cpa.message), findsOneWidget);
      expect(find.text('RFC'), findsNothing, reason: 'no RFC to review');

      await tapText(tester, 'Continuar');
      expect(find.byType(OnboardingScreen), findsNothing);
    });
  });

  group('step 5: tax inputs (module_tax_engine on)', () {
    Future<void> createMoral(WidgetTester tester) async {
      await openRegimes(
        tester,
        card: 'Persona moral',
        rfc: 'ABC680524P76',
        alsoEnabled: {AppModule.taxEngine},
      );
      await _pick(tester, '601');
      await tapText(tester, _next);
      await tapText(tester, _create);
    }

    testWidgets('offers the regime inputs as optional, then saves them', (
      tester,
    ) async {
      await createMoral(tester);

      expect(find.text('Paso 5 de 5'), findsOneWidget);
      expect(
        find.textContaining('Opcional. Con estos datos de 2026'),
        findsOneWidget,
      );
      await typeInto(tester, 'Coeficiente de utilidad', '0.2345');
      await tapText(tester, 'Guardar');

      final save = harness.repository.taxInputSaves.single;
      expect(save.regimeCode, '601');
      expect(save.fiscalYear, 2026);
      expect(save.values.single.key, 'profit_coefficient');
      expect(save.values.single.valueRate, '0.2345');
      expect(find.byType(OnboardingScreen), findsNothing);
    });

    testWidgets('Omitir skips without a call', (tester) async {
      await createMoral(tester);

      await tapText(tester, 'Omitir');

      expect(harness.repository.taxInputSaves, isEmpty);
      expect(find.byType(OnboardingScreen), findsNothing);
    });

    testWidgets('money inputs are read in centavos', (tester) async {
      harness.repository.catalogs[PersonType.moral] = [
        regime(
          '601',
          personType: PersonType.moral,
          requiredInputs: [inputDef('ytd_income', label: 'Ingresos del año')],
        ),
      ];
      await openRegimes(
        tester,
        card: 'Persona moral',
        rfc: 'ABC680524P76',
        alsoEnabled: {AppModule.taxEngine},
      );
      await _pick(tester, '601');
      await tapText(tester, _create);

      await typeInto(tester, 'Ingresos del año', '1,234.50');
      await tapText(tester, 'Guardar');

      expect(
        harness.repository.taxInputSaves.single.values.single.valueCents,
        123450,
      );
    });

    testWidgets('a failed save stays with a notice', (tester) async {
      harness.repository.taxInputsFailure = const ApiUnreachable();
      await createMoral(tester);

      await typeInto(tester, 'Coeficiente de utilidad', '0.2');
      await tapText(tester, 'Guardar');

      expect(find.textContaining('No hay conexión.'), findsOneWidget);
    });

    testWidgets('a kind this build does not know is left out', (tester) async {
      harness.repository.catalogs[PersonType.moral] = [
        regime(
          '601',
          personType: PersonType.moral,
          requiredInputs: [
            inputDef('x', label: 'Otro dato', kind: TaxInputDefKind.$unknown),
            inputDef('ytd', label: 'Ingresos del año'),
          ],
        ),
      ];
      await openRegimes(
        tester,
        card: 'Persona moral',
        rfc: 'ABC680524P76',
        alsoEnabled: {AppModule.taxEngine},
      );
      await _pick(tester, '601');
      await tapText(tester, _create);

      expect(find.text('Otro dato'), findsNothing);
      expect(find.text('Ingresos del año'), findsOneWidget);
    });
  });

  for (final brightness in Brightness.values) {
    testWidgets('steps 3 and 4 meet the tap-target, label and contrast '
        'guidelines (${brightness.name})', (tester) async {
      final semantics = tester.ensureSemantics();
      await openRegimes(tester, brightness: brightness);
      await _pick(tester, '625');

      await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
      await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
      await expectLater(tester, meetsGuideline(textContrastGuideline));

      await tapText(tester, _next);
      await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
      await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
      await expectLater(tester, meetsGuideline(textContrastGuideline));
      semantics.dispose();
    });
  }
}

bool _switchValue(WidgetTester tester, String label) => tester
    .widget<SwitchListTile>(
      find.ancestor(
        of: find.text(label),
        matching: find.byType(SwitchListTile),
      ),
    )
    .value;
