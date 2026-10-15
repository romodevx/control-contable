import 'package:control_contable/api/generated/models/counterparty_kind.dart';
import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/api/generated/models/withholder.dart';
import 'package:control_contable/core/access/read_only_access.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/incomes/incomes_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/choice_finders.dart';
import '../../helpers/withholding_fixtures.dart';
import '../taxpayers/pump_onboarding.dart' show inputLabeled, tapText, typeInto;
import '../taxpayers/taxpayers_fakes.dart';
import 'incomes_fakes.dart';
import 'pump_incomes.dart';

const _amount = 'Monto sin IVA';
const _description = 'Descripción (opcional)';
const _collectedOn = 'Fecha de cobro';
const _ivaAmount = 'IVA cobrado (opcional)';
const _more = 'Más detalles';
const _q1 = '¿A quién le facturaste?';
const _readOnlyTitle = 'Tu cuenta es de solo consulta';
const _logoId = 'i-logo';
const _invalid = 'Revisa este dato.';

String _textOf(WidgetTester tester, String label) =>
    tester.widget<TextField>(inputLabeled(label)).controller!.text;

void main() {
  late IncomesHarness harness;

  setUp(() => harness = IncomesHarness());

  /// Opens the list holding the US-I1 income, then taps its row.
  Future<void> openEdit(
    WidgetTester tester, {
    IvaRate ivaRate = IvaRate.rate16,
    int ivaCents = 160000,
    CounterpartyKind counterpartyKind = CounterpartyKind.pm,
  }) async {
    harness.incomes.incomes.add(
      income(
        id: _logoId,
        isrWithheldCents: 100000,
        ivaWithheldCents: 106667,
        ivaRate: ivaRate,
        ivaCents: ivaCents,
        counterpartyKind: counterpartyKind,
        withholding: withholding(
          isrCents: 100000,
          ivaCents: 106667,
          withholder: Withholder.counterparty,
          netCents: ivaCents + 1000000 - 206667,
        ),
      ),
    );
    await harness.pump(tester, taxpayer: anaOneRegime);
    await tester.tap(find.text('Diseño de logotipo'));
    await tester.pumpAndSettle();
  }

  group('M3-09 edit', () {
    testWidgets('tapping a row opens the same form page, prefilled, with one '
        'Guardar and no "Guardar y crear otro"', (tester) async {
      await openEdit(tester);

      expect(harness.location, IncomesRoutes.editOf(_logoId));
      expect(find.text('Editar ingreso'), findsOneWidget);
      expect(_textOf(tester, _collectedOn), '05/10/2026');
      expect(_textOf(tester, _amount), '10,000.00');
      expect(find.text('Guardar y crear otro'), findsNothing);
      // The stored derivation shows at once, read-only, without asking.
      expect(
        find.text(r'Te retienen: ISR $1,000.00 · IVA $1,066.67'),
        findsOneWidget,
      );
      expect(harness.incomes.previewRequests, isEmpty);
      // The stored facts preselect the questions.
      expect(chosenIn(tester, _q1), CounterpartyKind.pm.json);
      await tapText(tester, _more);
      expect(_textOf(tester, _description), 'Diseño de logotipo');
    });

    testWidgets('a record the server has not derived yet (withholding null) '
        'opens prefilled and asks for its derivation', (tester) async {
      harness.incomes.incomes.add(
        income(id: _logoId, isrWithheldCents: 100000, derived: false),
      );
      await harness.pump(tester, taxpayer: anaOneRegime);
      await tester.tap(find.text('Diseño de logotipo'));
      await tester.pumpAndSettle();
      await settlePreview(tester);

      expect(find.text('Editar ingreso'), findsOneWidget);
      expect(chosenIn(tester, _q1), CounterpartyKind.pm.json);
      expect(
        harness.incomes.previewRequests.single.counterpartyKind,
        CounterpartyKind.pm,
      );
    });

    testWidgets('changing who was billed re-derives it and sends the new '
        'facts (US-W9)', (tester) async {
      await openEdit(tester);

      await chooseIn(
        tester,
        _q1,
        'Una persona con negocio o profesión (persona física con actividad)',
      );
      await settlePreview(tester);
      expect(
        harness.incomes.previewRequests.single.counterpartyKind,
        CounterpartyKind.pfWithActivity,
      );
      expect(
        find.text('Nadie te retiene impuestos en esta operación.'),
        findsOneWidget,
      );
      await tapText(tester, 'Guardar');

      expect(harness.incomes.patches, [
        {'counterparty_kind': CounterpartyKind.pfWithActivity.json},
      ]);
    });

    testWidgets('the row is announced as a button that edits the income', (
      tester,
    ) async {
      final semantics = tester.ensureSemantics();
      harness.incomes.incomes.add(income(id: _logoId));
      await harness.pump(tester, taxpayer: anaOneRegime);

      expect(
        find.bySemanticsLabel(RegExp('Editar ingreso: Diseño de logotipo')),
        findsOneWidget,
      );
      semantics.dispose();
    });

    testWidgets('a change sends only the changed fields (IVA follows the '
        'amount) and returns to the list with the new total', (tester) async {
      await openEdit(tester);

      await typeInto(tester, _amount, '12000');
      await tapText(tester, 'Guardar');

      expect(harness.incomes.patches, [
        {'amount_cents': 1200000, 'iva_cents': 192000},
      ]);
      expect(harness.location, IncomesRoutes.list);
      expect(find.text('Ingreso actualizado'), findsOneWidget);
      expect(find.text(r'$13,920.00'), findsWidgets);
    });

    testWidgets('clearing an optional field sends null (merge-patch clears '
        'it)', (tester) async {
      await openEdit(tester);
      await tapText(tester, _more);

      await typeInto(tester, _description, '');
      await tapText(tester, 'Guardar');

      expect(harness.incomes.patches, [
        {'description': null},
      ]);
    });

    testWidgets('the stored rate, IVA and facts are never replaced by the '
        'defaults suggestions', (tester) async {
      harness.incomes.defaultsAnswer = noWithholdingDefaults.copyWith(
        lastUsed: lastUsed(counterpartyKind: CounterpartyKind.foreign),
      );
      await openEdit(tester, ivaRate: IvaRate.rate8, ivaCents: 80000);

      final rates = tester.widget<RadioGroup<String>>(
        find.ancestor(
          of: find.widgetWithText(RadioListTile<String>, '16 %'),
          matching: find.byType(RadioGroup<String>),
        ),
      );
      expect(rates.groupValue, IvaRate.rate8.json);
      await tapText(tester, 'Guardar');

      // Nothing changed: nothing is sent, the page closes.
      expect(harness.incomes.patches, isEmpty);
      expect(harness.location, IncomesRoutes.list);
    });

    testWidgets('a stored 16 % stays when the defaults now suggest another '
        'rate', (tester) async {
      harness.incomes.defaultsAnswer = moralDefaults612.copyWith(
        suggestedIvaRate: IvaRate.rate0,
      );
      await openEdit(tester);

      await tapText(tester, 'Guardar');

      expect(harness.incomes.patches, isEmpty);
    });

    testWidgets('a server verdict lands on its field', (tester) async {
      await openEdit(tester);
      harness.incomes.failNext['update'] = ApiValidationFailure([
        detail('amount_cents', 'must_be_positive'),
      ]);

      await typeInto(tester, _amount, '12000');
      await tapText(tester, 'Guardar');

      expect(harness.location, IncomesRoutes.editOf(_logoId));
      expect(find.text(_invalid), findsWidgets);
    });

    testWidgets('a read-only account gets the paywall and keeps the form', (
      tester,
    ) async {
      await openEdit(tester);
      harness.incomes.failNext['update'] = const ApiEntitlementRequired();

      await typeInto(tester, _amount, '12000');
      await tapText(tester, 'Guardar');

      expect(find.text(_readOnlyTitle), findsOneWidget);
      expect(harness.container.read(readOnlyAccessProvider), isTrue);
    });

    testWidgets('an income deleted elsewhere says so instead of saving', (
      tester,
    ) async {
      await openEdit(tester);
      harness.incomes.failNext['update'] = const ApiNotFound();

      await typeInto(tester, _amount, '12000');
      await tapText(tester, 'Guardar');

      expect(
        find.text('Este ingreso ya no existe: se eliminó en otro lugar.'),
        findsOneWidget,
      );
    });

    testWidgets('opened by its location, the income is loaded first', (
      tester,
    ) async {
      harness.incomes.incomes.add(income(id: _logoId));
      await harness.pump(
        tester,
        taxpayer: anaOneRegime,
        location: IncomesRoutes.editOf(_logoId),
      );

      expect(harness.incomes.calls, contains('get:$_logoId'));
      await tapText(tester, _more);
      expect(_textOf(tester, _description), 'Diseño de logotipo');
    });

    testWidgets('the IVA a user wrote stays theirs when the amount changes', (
      tester,
    ) async {
      await openEdit(tester, ivaCents: 150000);
      await tapText(tester, _more);
      expect(_textOf(tester, _ivaAmount), '1,500.00');

      await typeInto(tester, _amount, '12000');
      await tapText(tester, 'Guardar');

      expect(harness.incomes.patches, [
        {'amount_cents': 1200000},
      ]);
    });

    testWidgets('a business-rule refusal without a field is a form notice', (
      tester,
    ) async {
      await openEdit(tester);
      harness.incomes.failNext['update'] = ApiRejected(
        code: ErrorCode.businessRuleViolation,
        details: [detail('taxpayerId', 'taxpayer_archived')],
      );

      await typeInto(tester, _amount, '12000');
      await tapText(tester, 'Guardar');

      expect(harness.location, IncomesRoutes.editOf(_logoId));
      expect(
        find.text(
          'Este contribuyente está archivado. Desarchívalo para hacer cambios.',
        ),
        findsOneWidget,
      );
    });
  });
}
