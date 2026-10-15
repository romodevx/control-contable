import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/features/dashboard/data/dashboard_repository.dart';
import 'package:control_contable/features/taxpayers/data/active_taxpayer_store.dart';
import 'package:control_contable/features/taxpayers/data/taxpayers_repository.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_test/flutter_test.dart';

import '../auth/pump_auth.dart';
import '../dashboard/pump_dashboard.dart' show FakeDashboardRepository;
import '../taxpayers/taxpayers_fakes.dart';

/// An iPhone 17 in points.
const Size _iphone = Size(402, 874);

/// A small Android phone, the narrowest the app supports well.
const Size _smallPhone = Size(360, 740);

const String _name = 'Ana RESICO';

/// Every module that puts an action in Home's app bar, plus the dashboard
/// every preset shows under it (auth is the harness's own).
const Set<AppModule> _barModules = {
  AppModule.taxpayers,
  AppModule.taxEngine,
  AppModule.dashboard,
  AppModule.annual,
  AppModule.history,
  AppModule.reminders,
};

/// iOS QA walk 2026-10-15 (demo_full on an iPhone 17): the bar held the
/// switcher, four icons and a text "Cerrar sesión" that grew with the text
/// size: at 200 % the row overflowed, clipped "Cerrar sesión" and pushed
/// the switcher off screen (WCAG 1.4.4).
void main() {
  Future<void> pumpHome(WidgetTester tester, Size size) {
    tester.view
      ..physicalSize = size
      ..devicePixelRatio = 1;
    addTearDown(tester.view.reset);
    final taxpayers = FakeTaxpayersRepository();
    taxpayers.taxpayers.add(taxpayer(displayName: _name));
    return AuthHarness().pump(
      tester,
      signedIn: true,
      alsoEnabled: _barModules,
      overrides: [
        taxpayersRepositoryProvider.overrideWithValue(taxpayers),
        dashboardRepositoryProvider.overrideWithValue(
          FakeDashboardRepository(),
        ),
        activeTaxpayerStoreProvider.overrideWithValue(
          InMemoryActiveTaxpayerStore(),
        ),
      ],
    );
  }

  for (final MapEntry(key: device, value: size) in {
    'iPhone': _iphone,
    'small phone': _smallPhone,
  }.entries) {
    for (final scale in [1.0, 2.0]) {
      testWidgets('$device at ${(scale * 100).round()} % text: the bar fits '
          'and the switcher stays on screen', (tester) async {
        tester.platformDispatcher.textScaleFactorTestValue = scale;
        addTearDown(tester.platformDispatcher.clearTextScaleFactorTestValue);
        final semantics = tester.ensureSemantics();
        await pumpHome(tester, size);

        expect(tester.takeException(), isNull, reason: 'overflow');
        final switcher = tester.getRect(
          find.bySemanticsLabel(RegExp('^Contribuyente activo: $_name')),
        );
        expect(switcher.left, greaterThanOrEqualTo(0));
        expect(switcher.width, greaterThanOrEqualTo(48));
        semantics.dispose();
      });
    }
  }

  testWidgets('sign-out keeps its name for screen readers and as a tooltip', (
    tester,
  ) async {
    final semantics = tester.ensureSemantics();
    await pumpHome(tester, _iphone);

    // Like "Mi cuenta": iOS reads the tooltip as the button's label.
    final signOut = find.byTooltip('Cerrar sesión');
    expect(signOut, findsOneWidget);
    expect(tester.getSemantics(signOut).tooltip, 'Cerrar sesión');
    semantics.dispose();
  });
}
