import 'dart:ui';

import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/features/taxpayers/taxpayers_routes.dart';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'pump_onboarding.dart';
import 'taxpayers_fakes.dart';

final Taxpayer _ana = taxpayer();
final Taxpayer _taller = taxpayer(
  id: secondTaxpayerId,
  displayName: 'Taller SA',
  personType: PersonType.moral,
  rfcMasked: 'TAL******XY9',
  createdAt: DateTime.utc(2026, 10, 14),
);
final Taxpayer _viejo = taxpayer(
  id: archivedTaxpayerId,
  displayName: 'Viejo',
  rfcMasked: 'VIEJ******ZZ1',
  archivedAt: DateTime.utc(2026, 10, 12),
);

const _switcherLabel = 'Contribuyente activo: Ana. Cambiar';

Future<void> _openSwitcher(WidgetTester tester) async {
  await tester.tap(find.bySemanticsLabel(_switcherLabel));
  await tester.pumpAndSettle();
}

void main() {
  late OnboardingHarness harness;

  setUp(() => harness = OnboardingHarness());

  Future<void> pumpHome(
    WidgetTester tester, {
    List<Taxpayer>? taxpayers,
    bool enabled = true,
  }) => harness.pump(
    tester,
    location: null,
    enabled: enabled,
    taxpayers: taxpayers ?? [_ana, _taller, _viejo],
  );

  testWidgets('the top bar names the active taxpayer', (tester) async {
    await pumpHome(tester);

    expect(find.bySemanticsLabel(_switcherLabel), findsOneWidget);
    expect(harness.store.id, _ana.id);
  });

  testWidgets('module_taxpayers off: no switcher, nothing loaded, no '
      'onboarding', (tester) async {
    await pumpHome(tester, enabled: false, taxpayers: const []);

    expect(find.bySemanticsLabel(_switcherLabel), findsNothing);
    expect(harness.repository.calls, isEmpty);
    expect(harness.location, AppRoutes.home);
  });

  testWidgets('the sheet lists the non-archived taxpayers with masked RFC '
      'and person type; archived ones only on demand', (tester) async {
    await pumpHome(tester);
    await _openSwitcher(tester);

    expect(find.text('Taller SA'), findsOneWidget);
    expect(find.text('GOMA••••••AB6'), findsOneWidget);
    expect(find.text('TAL••••••XY9'), findsOneWidget);
    expect(find.text('Persona física'), findsOneWidget);
    expect(find.text('Persona moral'), findsOneWidget);
    expect(find.textContaining('*'), findsNothing);
    expect(find.text('Viejo'), findsNothing);

    await tapText(tester, 'Mostrar archivados');

    expect(find.text('Viejo'), findsOneWidget);
    expect(find.text('Archivado'), findsOneWidget);
  });

  testWidgets('the switcher and its sheet meet the tap-target, label and '
      'contrast guidelines', (tester) async {
    final handle = tester.ensureSemantics();
    await pumpHome(tester);
    await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
    await _openSwitcher(tester);

    await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
    await expectLater(tester, meetsGuideline(iOSTapTargetGuideline));
    await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
    await expectLater(tester, meetsGuideline(textContrastGuideline));
    handle.dispose();
  });

  testWidgets('the active taxpayer reads as selected', (tester) async {
    await pumpHome(tester);
    await _openSwitcher(tester);

    final ana = tester.getSemantics(
      find.ancestor(of: find.text('Ana'), matching: find.byType(ListTile)),
    );
    expect(ana.flagsCollection.isSelected, Tristate.isTrue);
  });

  testWidgets('picking another taxpayer makes it active and keeps it', (
    tester,
  ) async {
    await pumpHome(tester);
    await _openSwitcher(tester);

    await tapText(tester, 'Taller SA');

    expect(
      find.bySemanticsLabel('Contribuyente activo: Taller SA. Cambiar'),
      findsOneWidget,
    );
    expect(harness.store.id, secondTaxpayerId);
  });

  testWidgets('each row opens that taxpayer settings', (tester) async {
    await pumpHome(tester);
    await _openSwitcher(tester);

    await tester.tap(find.byTooltip('Ajustes de Taller SA'));
    await tester.pumpAndSettle();

    expect(harness.location, TaxpayersRoutes.settings(secondTaxpayerId));
  });

  testWidgets('a user with no taxpayer lands on the onboarding once', (
    tester,
  ) async {
    await pumpHome(tester, taxpayers: const []);

    expect(harness.location, TaxpayersRoutes.onboarding);

    await tester.tap(find.byTooltip('Atrás'));
    await tester.pumpAndSettle();

    expect(harness.location, AppRoutes.home);
    await tester.pumpAndSettle();
    expect(harness.location, AppRoutes.home, reason: 'never reopened');
  });

  testWidgets('only archived taxpayers: home stays, offering to register '
      'one', (tester) async {
    await pumpHome(tester, taxpayers: [_viejo]);

    expect(harness.location, AppRoutes.home);
    expect(find.text('Registrar contribuyente'), findsOneWidget);
  });

  testWidgets('a finished onboarding makes the new taxpayer active', (
    tester,
  ) async {
    await pumpHome(tester, taxpayers: [_taller]);
    harness.repository.createAnswers.add(
      taxpayer(id: testClientId, displayName: 'Nueva'),
    );
    await tapText(tester, 'Registrar contribuyente');
    await passIdentity(tester);
    await tapText(tester, '626 · RESICO');
    await tapText(tester, 'Crear contribuyente');

    expect(harness.location, AppRoutes.home);
    expect(harness.store.id, testClientId);
    expect(
      find.bySemanticsLabel('Contribuyente activo: Nueva. Cambiar'),
      findsOneWidget,
    );
  });
  group('the active taxpayer changed elsewhere (another device, the web)', () {
    Future<void> resume(WidgetTester tester) async {
      tester.binding
        ..handleAppLifecycleStateChanged(AppLifecycleState.inactive)
        ..handleAppLifecycleStateChanged(AppLifecycleState.resumed);
      await tester.pumpAndSettle();
    }

    testWidgets('deleted: on resume the next taxpayer becomes active', (
      tester,
    ) async {
      await pumpHome(tester);
      harness.repository.taxpayers.removeWhere((t) => t.id == _ana.id);

      await resume(tester);

      expect(
        find.bySemanticsLabel('Contribuyente activo: Taller SA. Cambiar'),
        findsOneWidget,
      );
      expect(harness.store.id, _taller.id);
    });

    testWidgets('archived: on resume the next taxpayer becomes active', (
      tester,
    ) async {
      await pumpHome(tester);
      harness.repository.taxpayers[0] = _ana.copyWith(
        archivedAt: DateTime.utc(2026, 10, 13),
      );

      await resume(tester);

      expect(harness.store.id, _taller.id);
    });

    testWidgets('a failed reload on resume keeps the last list', (
      tester,
    ) async {
      await pumpHome(tester);
      harness.repository.failNext['list'] = const ApiUnreachable();

      await resume(tester);

      expect(find.bySemanticsLabel(_switcherLabel), findsOneWidget);
    });
  });

  group('an active taxpayer with no regime in effect', () {
    const prompt = 'Agrega un régimen para ver tu estimado.';
    final ended = _ana.copyWith(
      regimes: [heldRegime('626', endPeriod: '2026-09')],
    );

    testWidgets('home stays and links to adding a regime, no onboarding', (
      tester,
    ) async {
      await pumpHome(tester, taxpayers: [ended]);

      expect(harness.location, AppRoutes.home);
      expect(find.text(prompt), findsOneWidget);

      await tapText(tester, 'Agregar régimen');

      expect(harness.location, TaxpayersRoutes.addRegime(_ana.id));
    });

    testWidgets('a regime ending this month still counts', (tester) async {
      await pumpHome(
        tester,
        taxpayers: [
          _ana.copyWith(regimes: [heldRegime('626', endPeriod: '2026-10')]),
        ],
      );

      expect(find.text(prompt), findsNothing);
    });

    testWidgets('an open regime: no prompt', (tester) async {
      await pumpHome(
        tester,
        taxpayers: [
          _ana.copyWith(regimes: [heldRegime('626')]),
        ],
      );

      expect(find.text(prompt), findsNothing);
    });
  });
}
