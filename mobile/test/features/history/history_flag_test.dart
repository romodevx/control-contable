import 'dart:async';

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
/// this is UX through the single FlagGate.
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

  testWidgets('implied OFF: with the engine OFF the server leaves '
      'module_history out, so no entry shows', (tester) async {
    await harness.pump(
      tester,
      modules: homeModules.difference({AppModule.taxEngine}),
    );

    expect(historyEntry(), findsNothing);
  });
}
