import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/core/access/read_only_access.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/router/app_router.dart';
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
const _archivedBanner =
    'Este contribuyente está archivado. Puedes consultarlo; '
    'para hacer cambios, desarchívalo.';

void main() {
  late OnboardingHarness harness;

  setUp(() => harness = OnboardingHarness());

  Future<void> pumpSettings(WidgetTester tester, {Taxpayer? of}) {
    final shown = of ?? _ana;
    return harness.pump(
      tester,
      location: TaxpayersRoutes.settings(shown.id),
      taxpayers: [
        shown,
        taxpayer(id: secondTaxpayerId, displayName: 'B'),
      ],
    );
  }

  testWidgets('shows the name, the masked RFC, the person type and every '
      'regime with its period', (tester) async {
    await pumpSettings(tester);

    expect(find.text('Ana'), findsOneWidget);
    expect(find.text('GOMA••••••AB6'), findsOneWidget);
    expect(find.text('Persona física'), findsOneWidget);
    expect(find.text('612 · Régimen 612'), findsOneWidget);
    expect(find.text('Desde enero de 2026'), findsNWidgets(2));
    expect(find.text('De enero de 2026 a junio de 2026'), findsOneWidget);
    expect(find.text('Terminado'), findsOneWidget);
  });

  for (final brightness in Brightness.values) {
    testWidgets('meets the tap-target, label and contrast guidelines '
        '(${brightness.name})', (tester) async {
      final handle = tester.ensureSemantics();
      await harness.pump(
        tester,
        location: TaxpayersRoutes.settings(_id),
        taxpayers: [_ana],
        brightness: brightness,
      );

      await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
      await expectLater(tester, meetsGuideline(iOSTapTargetGuideline));
      await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
      await expectLater(tester, meetsGuideline(textContrastGuideline));
      handle.dispose();
    });
  }

  testWidgets('module_taxpayers off: the settings are not found, nothing '
      'loads', (tester) async {
    await harness.pump(
      tester,
      enabled: false,
      location: TaxpayersRoutes.settings(_id),
      taxpayers: [_ana],
    );

    expect(harness.location, AppRoutes.notFound);
    expect(harness.repository.calls, isEmpty);
  });

  testWidgets('a taxpayer that is gone says so', (tester) async {
    await harness.pump(tester, location: TaxpayersRoutes.settings('gone'));

    expect(find.text('Este contribuyente ya no existe.'), findsOneWidget);
  });

  group('rename (inline row)', () {
    testWidgets('saves the new name in place', (tester) async {
      await pumpSettings(tester);
      await tapText(tester, 'Editar nombre');

      await typeInto(tester, 'Nombre para identificarlo', 'Ana G');
      await tapText(tester, 'Guardar');

      expect(harness.repository.calls, contains('rename:Ana G'));
      expect(find.text('Ana G'), findsWidgets);
      expect(find.text('Guardamos el nombre.'), findsOneWidget);
    });

    testWidgets('an empty name is refused on submit, nothing sent', (
      tester,
    ) async {
      await pumpSettings(tester);
      await tapText(tester, 'Editar nombre');

      await typeInto(tester, 'Nombre para identificarlo', '  ');
      await tapText(tester, 'Guardar');

      expect(find.text('Escribe este dato.'), findsWidgets);
      expect(
        harness.repository.calls.where((c) => c.startsWith('rename')),
        isEmpty,
      );
    });

    testWidgets('Cancelar leaves the name as it was', (tester) async {
      await pumpSettings(tester);
      await tapText(tester, 'Editar nombre');
      await typeInto(tester, 'Nombre para identificarlo', 'Otro');

      await tapText(tester, 'Cancelar');

      expect(find.text('Ana'), findsOneWidget);
      expect(find.text('Editar nombre'), findsOneWidget);
    });
  });

  group('RFC correction', () {
    testWidgets('sends the RFC normalized and confirms', (tester) async {
      await pumpSettings(tester);
      await tapText(tester, 'Corregir RFC');

      await typeInto(tester, 'RFC', 'goma-800101 ab6');
      await tapText(tester, 'Guardar');

      expect(harness.repository.corrections.single.rfc, 'GOMA800101AB6');
      expect(find.text('Corregimos el RFC.'), findsOneWidget);
    });

    testWidgets('a check-digit mismatch is saved and asks to double-check', (
      tester,
    ) async {
      harness.repository.correctAnswers.add(
        taxpayer(
          warnings: const [
            Warning(
              code: 'rfc_check_digit_mismatch',
              field: 'rfc',
              message: 'x',
            ),
          ],
        ),
      );
      await pumpSettings(tester);
      await tapText(tester, 'Corregir RFC');

      await typeInto(tester, 'RFC', 'GOMA800101AB7');
      await tapText(tester, 'Guardar');

      expect(find.text('Revisa que tu RFC esté bien escrito'), findsOneWidget);
      expect(find.textContaining('El último carácter'), findsOneWidget);
    });

    testWidgets('a wrong length is caught before sending', (tester) async {
      await pumpSettings(tester);
      await tapText(tester, 'Corregir RFC');

      await typeInto(tester, 'RFC', 'GOMA80010');
      await tapText(tester, 'Guardar');

      expect(
        find.text('El RFC de una persona física tiene 13 caracteres.'),
        findsWidgets,
      );
      expect(harness.repository.corrections, isEmpty);
    });

    testWidgets('an RFC already registered shows on the field', (tester) async {
      harness.repository.correctAnswers.add(
        const ApiRejected(code: ErrorCode.rfcAlreadyRegistered),
      );
      await pumpSettings(tester);
      await tapText(tester, 'Corregir RFC');

      await typeInto(tester, 'RFC', 'GOMA800101AB6');
      await tapText(tester, 'Guardar');

      expect(find.textContaining('Ya registraste'), findsWidgets);
    });
  });

  group('read-only user (OQ-8)', () {
    testWidgets('a refused write shows the paywall; later writes open it '
        'without asking the server', (tester) async {
      harness.repository.failNext['rename'] = const ApiEntitlementRequired();
      await pumpSettings(tester);
      await tapText(tester, 'Editar nombre');
      await typeInto(tester, 'Nombre para identificarlo', 'Ana G');
      await tapText(tester, 'Guardar');

      expect(find.text(_paywallTitle), findsOneWidget);
      expect(harness.container.read(readOnlyAccessProvider), isTrue);
      await tapText(tester, 'Entendido');

      await tapText(tester, 'Corregir RFC');

      expect(find.text(_paywallTitle), findsOneWidget);
      expect(harness.repository.corrections, isEmpty);
    });

    testWidgets('archive still works; its undo needs a plan', (tester) async {
      harness.repository.failNext['unarchive'] = const ApiEntitlementRequired();
      await pumpSettings(tester);

      await tapText(tester, 'Archivar');
      await tapText(tester, 'Deshacer');

      expect(harness.repository.calls, contains('archive:$_id'));
      expect(find.text(_paywallTitle), findsOneWidget);
      expect(find.text(_archivedBanner), findsOneWidget);
    });

    // QA M2-12 D2: unarchive is a `402` for a read-only user (OQ-8), so
    // once that is known an archive is one-way: confirm first (house rule:
    // confirmation for the irreversible), and no Undo that only paywalls.
    testWidgets('known read-only: archiving asks first and offers no Undo', (
      tester,
    ) async {
      Iterable<String> archives() =>
          harness.repository.calls.where((c) => c.startsWith('archive:'));
      await pumpSettings(tester);
      harness.container.read(readOnlyAccessProvider.notifier).markReadOnly();
      await tester.pump();

      await tapText(tester, 'Archivar');

      expect(find.text('¿Archivar a «Ana»?'), findsOneWidget);
      expect(
        find.textContaining('no podrás desarchivarlo hasta tener un plan'),
        findsOneWidget,
      );
      expect(archives(), isEmpty);
      await tapText(tester, 'Cancelar');
      expect(archives(), isEmpty);
      expect(find.text(_archivedBanner), findsNothing);

      await tapText(tester, 'Archivar');
      await tapText(tester, 'Archivar de todos modos');

      expect(archives(), ['archive:$_id']);
      expect(find.text(_archivedBanner), findsOneWidget);
      expect(find.text('Archivaste a «Ana».'), findsOneWidget);
      expect(find.text('Deshacer'), findsNothing);
    });
  });

  group('archive (reversible: do it, then offer Undo)', () {
    testWidgets('archives at once and keeps the Undo until dismissed', (
      tester,
    ) async {
      await pumpSettings(tester);

      await tapText(tester, 'Archivar');

      expect(harness.repository.calls, contains('archive:$_id'));
      expect(find.text(_archivedBanner), findsOneWidget);
      expect(find.text('Editar nombre'), findsNothing);
      expect(find.text('Desarchivar'), findsOneWidget);
      await tester.pump(const Duration(minutes: 5));
      expect(find.text('Archivaste a «Ana».'), findsOneWidget);
      expect(
        find.ancestor(
          of: find.text('Archivaste a «Ana».'),
          matching: find.byType(SnackBar),
        ),
        findsOneWidget,
      );

      await tapText(tester, 'Deshacer');

      expect(harness.repository.calls, contains('unarchive:$_id'));
      expect(find.text(_archivedBanner), findsNothing);
    });

    testWidgets('Undo still works after leaving the settings', (tester) async {
      await pumpSettings(tester);
      await tapText(tester, 'Archivar');

      harness.container.read(appRouterProvider).go(AppRoutes.home);
      await tester.pumpAndSettle();
      await tapText(tester, 'Deshacer');

      expect(tester.takeException(), isNull);
      expect(harness.repository.calls, contains('unarchive:$_id'));
      expect(find.text('«Ana» volvió a tu lista.'), findsOneWidget);
    });

    testWidgets('Undo needing a plan after leaving the settings says why', (
      tester,
    ) async {
      harness.repository.failNext['unarchive'] = const ApiEntitlementRequired();
      await pumpSettings(tester);
      await tapText(tester, 'Archivar');

      harness.container.read(appRouterProvider).go(AppRoutes.home);
      await tester.pumpAndSettle();
      await tapText(tester, 'Deshacer');

      expect(tester.takeException(), isNull);
      expect(
        find.textContaining('hacer cambios requiere un plan'),
        findsOneWidget,
      );
    });

    testWidgets('an archived taxpayer can be unarchived', (tester) async {
      await pumpSettings(
        tester,
        of: _ana.copyWith(archivedAt: DateTime.utc(2026, 10, 12)),
      );
      expect(find.text(_archivedBanner), findsOneWidget);

      await tapText(tester, 'Desarchivar');

      expect(harness.repository.calls, contains('unarchive:$_id'));
      expect(find.text('«Ana» volvió a tu lista.'), findsOneWidget);
    });
  });

  group('delete (irreversible: confirm first)', () {
    testWidgets('the dialog names the taxpayer and what is lost; Cancelar '
        'deletes nothing', (tester) async {
      await pumpSettings(tester);

      await tapText(tester, 'Eliminar');

      expect(find.text('¿Eliminar a «Ana»?'), findsOneWidget);
      expect(
        find.textContaining('Se borrarán sus ingresos y gastos'),
        findsOneWidget,
      );
      await tapText(tester, 'Cancelar');
      expect(
        harness.repository.calls.where((c) => c.startsWith('delete')),
        isEmpty,
      );
    });

    testWidgets('confirming deletes it and returns home', (tester) async {
      await pumpSettings(tester);

      await tapText(tester, 'Eliminar');
      await tapText(tester, 'Eliminar definitivamente');

      expect(harness.repository.calls, contains('delete:$_id'));
      expect(harness.location, AppRoutes.home);
      expect(find.text('Eliminaste a «Ana».'), findsOneWidget);
      expect(harness.store.id, secondTaxpayerId);
    });
  });
}
