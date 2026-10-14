import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/api/generated/models/warning_code.dart';
import 'package:control_contable/core/access/read_only_access.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/taxpayers/taxpayers_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'pump_onboarding.dart';
import 'taxpayers_fakes.dart';

final Taxpayer _ana = taxpayer(
  regimes: [
    heldRegime('612', options: {'iva_fully_taxable': true}),
    heldRegime('606', options: {'deduction_mode': 'blind_35'}),
    heldRegime('626', endPeriod: '2026-06'),
  ],
);
final String _id = _ana.id;

const _paywallTitle = 'Tu cuenta es de solo consulta';

/// The card of the held regime [code].
Finder _regimeCard(String code) => find
    .ancestor(
      of: find.text('$code · Régimen $code'),
      matching: find.byType(Card),
    )
    .first;

Future<void> _tapIn(WidgetTester tester, Finder card, String text) async {
  final target = find.descendant(of: card, matching: find.text(text));
  await tester.ensureVisible(target);
  await tester.pumpAndSettle();
  await tester.tap(target);
  await tester.pumpAndSettle();
}

Finder _catalogRow(String code) => find
    .ancestor(
      of: find.textContaining('$code ·'),
      matching: find.byType(InkWell),
    )
    .first;

Future<void> _pick(WidgetTester tester, String code) async {
  await tester.ensureVisible(_catalogRow(code));
  await tester.pumpAndSettle();
  await tester.tap(_catalogRow(code));
  await tester.pumpAndSettle();
}

void main() {
  late OnboardingHarness harness;

  setUp(() => harness = OnboardingHarness());

  Future<void> pumpSettings(WidgetTester tester, {Taxpayer? of}) =>
      harness.pump(
        tester,
        location: TaxpayersRoutes.settings((of ?? _ana).id),
        taxpayers: [of ?? _ana],
      );

  group('end a regime (effective period)', () {
    testWidgets('ends it in the current month by default', (tester) async {
      await pumpSettings(tester);
      await _tapIn(tester, _regimeCard('606'), 'Terminar régimen');

      expect(find.text('Último mes en este régimen'), findsOneWidget);
      expect(find.text('octubre de 2026'), findsOneWidget);
      await _tapIn(tester, _regimeCard('606'), 'Terminar');

      expect(
        harness.repository.calls,
        contains('endRegime:606 2026-10 confirm=false'),
      );
      expect(find.text('Terminamos el régimen 606.'), findsOneWidget);
      expect(find.text('Terminado'), findsNWidgets(2));
    });

    testWidgets('another month can be chosen', (tester) async {
      await pumpSettings(tester);
      await _tapIn(tester, _regimeCard('606'), 'Terminar régimen');

      await tester.tap(find.text('octubre de 2026'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('septiembre de 2026').last);
      await tester.pumpAndSettle();
      await _tapIn(tester, _regimeCard('606'), 'Terminar');

      expect(
        harness.repository.calls,
        contains('endRegime:606 2026-09 confirm=false'),
      );
    });

    testWidgets('a regime with records asks to keep them first', (
      tester,
    ) async {
      harness.repository.failNext['endRegime'] = const ApiRejected(
        code: ErrorCode.regimeHasRecords,
      );
      await pumpSettings(tester);
      await _tapIn(tester, _regimeCard('606'), 'Terminar régimen');
      await _tapIn(tester, _regimeCard('606'), 'Terminar');

      expect(find.textContaining('tiene ingresos o gastos'), findsOneWidget);
      await _tapIn(
        tester,
        _regimeCard('606'),
        'Terminar y conservar registros',
      );

      expect(
        harness.repository.calls,
        contains('endRegime:606 2026-10 confirm=true'),
      );
    });

    testWidgets('records after the month are explained', (tester) async {
      harness.repository.failNext['endRegime'] = const ApiRejected(
        code: ErrorCode.businessRuleViolation,
        details: [
          api.FieldError(
            field: 'end_period',
            code: 'records_after_end_period',
            message: 'x',
          ),
        ],
      );
      await pumpSettings(tester);
      await _tapIn(tester, _regimeCard('606'), 'Terminar régimen');
      await _tapIn(tester, _regimeCard('606'), 'Terminar');

      expect(
        find.textContaining('Hay registros después de ese mes'),
        findsOneWidget,
      );
    });

    testWidgets('the only active regime cannot be ended', (tester) async {
      await pumpSettings(tester, of: taxpayer(regimes: [heldRegime('612')]));

      expect(find.text('Terminar régimen'), findsNothing);
      expect(find.textContaining('Es su único régimen activo'), findsOneWidget);
    });
  });

  group('regime options (OQ-7)', () {
    testWidgets('edits the options, warning they apply to every period', (
      tester,
    ) async {
      await pumpSettings(tester);
      await _tapIn(tester, _regimeCard('606'), 'Editar opciones');

      expect(harness.location, TaxpayersRoutes.regimeOptions(_id, '606'));
      expect(
        find.textContaining('aplica a todos los meses de este régimen'),
        findsOneWidget,
      );
      await tapText(tester, 'Gastos reales con factura');
      await tapText(tester, 'Guardar');

      expect(
        harness.repository.calls,
        contains('updateOptions:606 {deduction_mode: actual}'),
      );
      expect(harness.location, TaxpayersRoutes.settings(_id));
      expect(find.text('Guardamos las opciones.'), findsOneWidget);
    });

    testWidgets('a read-only user sees the paywall', (tester) async {
      harness.repository.failNext['updateOptions'] =
          const ApiEntitlementRequired();
      await pumpSettings(tester);
      await _tapIn(tester, _regimeCard('606'), 'Editar opciones');

      await tapText(tester, 'Guardar');

      expect(find.text(_paywallTitle), findsOneWidget);
    });
  });

  group('add a regime (wizard steps 3–4)', () {
    testWidgets('offers only regimes not active, warns about a conflict with '
        'a held one, and adds one without options', (tester) async {
      await pumpSettings(tester);
      await tapText(tester, 'Agregar régimen');

      expect(harness.location, TaxpayersRoutes.addRegime(_id));
      expect(find.textContaining('612 ·'), findsNothing);
      expect(find.textContaining('606 ·'), findsNothing);
      await _pick(tester, '626');
      expect(
        find.text(
          'No se puede combinar con el 612: El RESICO sustituye a este '
          'régimen.',
        ),
        findsOneWidget,
      );

      await _pick(tester, '605');
      await tapText(tester, 'Agregar régimen');

      expect(harness.repository.calls, contains('addRegime:605'));
      expect(harness.location, TaxpayersRoutes.settings(_id));
      expect(find.text('Agregamos el régimen 605.'), findsOneWidget);
      expect(find.text('605 · Régimen 605'), findsOneWidget);
    });

    testWidgets('a regime with options goes through its options page', (
      tester,
    ) async {
      await pumpSettings(tester);
      await tapText(tester, 'Agregar régimen');
      await _pick(tester, '621');
      await tapText(tester, 'Siguiente');

      await tapText(tester, 'Año 2');
      await tapText(tester, 'Agregar régimen');

      expect(harness.repository.calls, contains('addRegime:621'));
      final added = harness.repository.taxpayers.single.regimes.last;
      expect(added.options, {'rif_year': '2'});
    });

    testWidgets('a pair under CPA review is added and explained', (
      tester,
    ) async {
      harness.repository.warnNext['addRegime'] = const [
        Warning(
          code: WarningCode.regimeCombinationNeedsCpaReview,
          field: 'regime_code',
          message: 'Pendiente de revisión.',
        ),
      ];
      await pumpSettings(tester);
      await tapText(tester, 'Agregar régimen');
      await _pick(tester, '605');
      await tapText(tester, 'Agregar régimen');

      expect(
        find.text(
          'Agregamos el régimen 605. Un contador debe confirmar la '
          'combinación: Pendiente de revisión.',
        ),
        findsOneWidget,
      );
    });

    testWidgets('a read-only user gets the paywall instead', (tester) async {
      harness.repository.failNext['addRegime'] = const ApiEntitlementRequired();
      await pumpSettings(tester);
      await tapText(tester, 'Agregar régimen');
      await _pick(tester, '605');
      await tapText(tester, 'Agregar régimen');

      expect(find.text(_paywallTitle), findsOneWidget);
      expect(harness.container.read(readOnlyAccessProvider), isTrue);
    });

    testWidgets('a known read-only user sees the paywall before choosing', (
      tester,
    ) async {
      await pumpSettings(tester);
      harness.container.read(readOnlyAccessProvider.notifier).markReadOnly();

      await tapText(tester, 'Agregar régimen');

      expect(find.text(_paywallTitle), findsOneWidget);
      expect(harness.repository.catalogCalls, isEmpty);
    });
  });
}
