import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/features/auth/application/pending_sign_up.dart';
import 'package:control_contable/features/auth/auth_routes.dart';
import 'package:control_contable/features/taxpayers/data/active_taxpayer_store.dart';
import 'package:control_contable/features/taxpayers/data/taxpayers_repository.dart';
import 'package:control_contable/features/taxpayers/taxpayers_routes.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/synthetic_password.dart';
import '../auth/pump_auth.dart';
import 'taxpayers_fakes.dart';

/// US-T1 with Home as the dashboard (M6) or the previous Home: a brand-new
/// account (sign-up, then the code) has no taxpayer, so the onboarding opens
/// by itself, also after another account signed out on the device. Pinned
/// in M6-06 because Home's body moved behind module_dashboard.
void main() {
  late AuthHarness harness;
  late _ServerTaxpayers taxpayers;

  setUp(() {
    harness = AuthHarness();
    taxpayers = _ServerTaxpayers();
  });

  Future<void> pump(
    WidgetTester tester, {
    required bool dashboard,
    bool signedIn = false,
  }) => harness.pump(
    tester,
    signedIn: signedIn,
    alsoEnabled: {AppModule.taxpayers, if (dashboard) AppModule.dashboard},
    overrides: [
      taxpayersRepositoryProvider.overrideWithValue(taxpayers),
      activeTaxpayerStoreProvider.overrideWithValue(
        InMemoryActiveTaxpayerStore(),
      ),
    ],
  );

  /// Registers a new account and enters its code, as the screens do.
  Future<void> signUpAndVerify(WidgetTester tester) async {
    harness.container
        .read(pendingSignUpProvider.notifier)
        .start(
          SignUpDraft(
            email: 'ana@example.mx',
            password: syntheticPassword(),
            name: 'Ana',
          ),
        );
    harness.container.read(appRouterProvider).go(AuthRoutes.verifyEmail);
    await tester.pumpAndSettle();
    await tester.enterText(inputLabeled('Código de verificación'), '482913');
    await tester.pumpAndSettle();
    expect(harness.tokens.token, 'refresh', reason: 'signed in');
  }

  void expectOnboarding() {
    expect(harness.router.state.uri.toString(), TaxpayersRoutes.onboarding);
    expect(find.text('Nuevo contribuyente'), findsOneWidget);
  }

  for (final dashboard in [true, false]) {
    final flag = 'module_dashboard ${dashboard ? 'on' : 'off'}';

    testWidgets('first account on the device: the onboarding opens ($flag)', (
      tester,
    ) async {
      await pump(tester, dashboard: dashboard);

      await signUpAndVerify(tester);

      expectOnboarding();
    });

    testWidgets('after another account signed out: the onboarding opens '
        'for the new one ($flag)', (tester) async {
      await pump(tester, dashboard: dashboard, signedIn: true);
      expectOnboarding(); // the first account, with no taxpayer either
      harness.router.pop();
      await tester.pumpAndSettle();
      taxpayers.signedOut = true;
      await tester.tap(find.text('Cerrar sesión'));
      await tester.pumpAndSettle();
      expect(harness.tokens.token, isNull, reason: 'signed out');
      taxpayers.signedOut = false;

      await signUpAndVerify(tester);

      expectOnboarding();
    });
  }
}

/// The taxpayers server as the device meets it: once the session is gone
/// ([signedOut]) the list answers 401 (a sign-out reloads it while Home is
/// still on screen).
class _ServerTaxpayers extends FakeTaxpayersRepository {
  /// Whether the device holds no session.
  bool signedOut = false;

  @override
  Future<List<Taxpayer>> listTaxpayers() {
    if (signedOut) return Future.error(const ApiUnauthorized());
    return super.listTaxpayers();
  }
}
