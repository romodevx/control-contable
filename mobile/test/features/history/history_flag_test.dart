import 'dart:async';

import 'package:control_contable/core/fiscal/tax_data_revision.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/features/history/history_routes.dart';
import 'package:control_contable/features/history/presentation/history_screen.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_test/flutter_test.dart';

import '../dashboard/pump_dashboard.dart';

/// `module_history` on the dashboard Home the presets show (M8-06):
/// `demo_core` / `demo_monetization` leave it out, `demo_full` holds it
/// (MODULE-history §3, OQ-H6). The server's 404 is the security boundary;
/// this is UX through the single FlagGate. Implied OFF (engine OFF)
/// is the server's: GET /v1/flags omits the flag, which the OFF case covers;
/// the preset smoke in backend/cmd/api proves the omission.
void main() {
  late DashboardHarness harness;

  setUp(() => harness = DashboardHarness());

  Finder historyEntry() => find.byTooltip('Historial');

  testWidgets('OFF (demo_core): the dashboard Home offers no "Historial" '
      'and the location is not History', (tester) async {
    await harness.pump(tester);

    expect(find.byKey(const ValueKey('dashboard-hero')), findsOneWidget);
    expect(historyEntry(), findsNothing);

    unawaited(
      harness.container.read(appRouterProvider).push<void>(HistoryRoutes.list),
    );
    await tester.pumpAndSettle();
    expect(find.byType(HistoryScreen), findsNothing);
  });

  testWidgets('ON (demo_full): the dashboard Home offers "Historial"', (
    tester,
  ) async {
    await harness.pump(tester, modules: {...homeModules, AppModule.history});

    expect(find.byKey(const ValueKey('dashboard-hero')), findsOneWidget);
    expect(historyEntry(), findsOneWidget);
  });

  // Device walk 2026-10-15 (F2): records saved from inside History (its
  // "Registrar ingreso", "Ver ingresos" → edit) left Home's dashboard
  // stale ("Aún no hay ingresos…", the trial count unspent) on return.
  // The record writes bump the tax-data revision (tax_data_revision_test).
  testWidgets('a write made over History reads the dashboard again', (
    tester,
  ) async {
    await harness.pump(tester, modules: {...homeModules, AppModule.history});
    expect(harness.repository.calls, hasLength(1));

    await tester.tap(historyEntry());
    await tester.pumpAndSettle();
    harness.container.read(taxDataRevisionProvider.notifier).bump();
    await harness.back(tester);

    expect(harness.repository.calls, hasLength(2));
  });

  testWidgets('History opened and closed without a write reads nothing', (
    tester,
  ) async {
    await harness.pump(tester, modules: {...homeModules, AppModule.history});

    await tester.tap(historyEntry());
    await tester.pumpAndSettle();
    await harness.back(tester);

    expect(harness.repository.calls, hasLength(1));
  });

  // M8-07 QA: the entry lives in Home's app bar, so it must follow
  // module_history alone, also on the placeholder Home (dashboard OFF).
  final placeholderHome = homeModules.difference({AppModule.dashboard});

  testWidgets('OFF with the dashboard OFF: the placeholder Home offers no '
      '"Historial"', (tester) async {
    await harness.pump(tester, modules: placeholderHome);

    expect(find.byKey(const ValueKey('dashboard-hero')), findsNothing);
    expect(historyEntry(), findsNothing);
  });

  testWidgets('ON with the dashboard OFF: the placeholder Home offers '
      '"Historial" and it opens History', (tester) async {
    await harness.pump(
      tester,
      modules: {...placeholderHome, AppModule.history},
    );

    expect(find.byKey(const ValueKey('dashboard-hero')), findsNothing);
    await tester.tap(historyEntry());
    await tester.pumpAndSettle();
    expect(find.byType(HistoryScreen), findsOneWidget);
  });
}
