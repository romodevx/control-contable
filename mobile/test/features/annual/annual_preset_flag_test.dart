import 'dart:async';

import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/features/annual/annual_routes.dart';
import 'package:control_contable/features/annual/presentation/annual_inputs_block.dart';
import 'package:control_contable/features/annual/presentation/annual_screen.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_test/flutter_test.dart';

import '../dashboard/pump_dashboard.dart';

/// `module_annual` on the dashboard Home the presets show (ANN-22):
/// `demo_core` / `demo_monetization` leave it out, `demo_full` holds it
/// (MODULE-annual §6, OQ-A7). The server's 404 is the security boundary;
/// this is UX through the single FlagGate (`useFlag`). Implied OFF (engine
/// or taxpayers OFF) is the server's: GET /v1/flags omits the flag, which
/// the OFF case covers; the preset smoke in backend/cmd/api proves the
/// omission.
void main() {
  late DashboardHarness harness;

  setUp(() => harness = DashboardHarness());

  Finder annualEntry() => find.byTooltip('Declaración anual');

  testWidgets('OFF (demo_core): Home offers no "Declaración anual" nor '
      '"Datos para tu anual", and the annual estimate is unreachable', (
    tester,
  ) async {
    await harness.pump(tester);

    expect(find.byKey(const ValueKey('dashboard-hero')), findsOneWidget);
    expect(annualEntry(), findsNothing);
    expect(find.text('Datos para tu anual'), findsNothing);

    unawaited(
      harness.container
          .read(appRouterProvider)
          .push<void>(AnnualRoutes.estimate),
    );
    await tester.pumpAndSettle();
    expect(find.byType(AnnualScreen), findsNothing);
    expect(find.byType(AnnualInputsBlock), findsNothing);
    expect(find.text('Datos para tu anual'), findsNothing);
  });

  testWidgets('ON (demo_full): Home offers "Declaración anual"', (
    tester,
  ) async {
    await harness.pump(
      tester,
      modules: {...homeModules, AppModule.history, AppModule.annual},
    );

    expect(find.byKey(const ValueKey('dashboard-hero')), findsOneWidget);
    expect(annualEntry(), findsOneWidget);
  });
}
