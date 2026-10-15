import 'dart:async';

import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/features/reminders/presentation/reminders_screen.dart';
import 'package:control_contable/features/reminders/reminders_routes.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_test/flutter_test.dart';

import '../dashboard/pump_dashboard.dart';

/// `module_reminders` on the dashboard Home the presets show (M9-11):
/// `demo_core` / `demo_monetization` leave it out, `demo_full` holds it
/// (MODULE-reminders §3, OQ-R8). The server's 404 is the security boundary;
/// this is UX through the single FlagGate (`useFlag`). Implied OFF
/// (taxpayers OFF) is the server's: GET /v1/flags omits the flag, which the
/// OFF case covers; the preset smoke in backend/cmd/api proves the omission.
void main() {
  late DashboardHarness harness;

  setUp(() => harness = DashboardHarness());

  Finder remindersEntry() => find.byTooltip('Recordatorios');

  testWidgets('OFF (demo_core): Home offers no "Recordatorios" and the '
      'reminder settings are unreachable', (tester) async {
    await harness.pump(tester);

    expect(find.byKey(const ValueKey('dashboard-hero')), findsOneWidget);
    expect(remindersEntry(), findsNothing);

    unawaited(
      harness.container
          .read(appRouterProvider)
          .push<void>(RemindersRoutes.settings),
    );
    await tester.pumpAndSettle();
    expect(find.byType(RemindersScreen), findsNothing);
  });

  testWidgets('ON (demo_full): Home offers "Recordatorios"', (tester) async {
    await harness.pump(tester, modules: {...homeModules, AppModule.reminders});

    expect(find.byKey(const ValueKey('dashboard-hero')), findsOneWidget);
    expect(remindersEntry(), findsOneWidget);
  });
}
