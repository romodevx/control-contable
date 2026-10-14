import 'dart:async';

import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/auth/auth_routes.dart';
import 'package:control_contable/features/auth/presentation/login_screen.dart';
import 'package:control_contable/features/home/presentation/home_screen.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';

import '../../helpers/synthetic_password.dart';
import 'auth_fakes.dart';
import 'pump_auth.dart';

const _incomesLocation = '/ingresos-prueba';
const _signOut = 'Cerrar sesión';

final FeatureModule _incomesFeature = FeatureModule(
  module: AppModule.incomes,
  routes: [
    GoRoute(
      path: _incomesLocation,
      builder: (context, state) => const Text('pantalla de ingresos'),
    ),
  ],
);

void main() {
  late AuthHarness harness;

  setUp(() => harness = AuthHarness());

  GoRouter router() => harness.container.read(appRouterProvider);

  Finder signOut() => find.widgetWithText(TextButton, _signOut);

  group('redirect guards (module_auth on)', () {
    testWidgets('signed out: a module location goes to login, then back', (
      tester,
    ) async {
      await harness.pump(
        tester,
        location: _incomesLocation,
        alsoEnabled: {AppModule.incomes},
        extraFeatures: [_incomesFeature],
      );
      expect(find.byType(LoginScreen), findsOneWidget);

      await tester.enterText(inputLabeled('Correo electrónico'), 'a@b.mx');
      await tester.enterText(inputLabeled('Contraseña'), syntheticPassword());
      await tester.tap(find.widgetWithText(FilledButton, 'Entrar'));
      await tester.pumpAndSettle();

      expect(find.text('pantalla de ingresos'), findsOneWidget);
    });

    testWidgets('signed in (stored refresh token): opens on home', (
      tester,
    ) async {
      await harness.pump(tester, signedIn: true);

      expect(find.byType(HomeScreen), findsOneWidget);
      expect(signOut(), findsOneWidget);
    });

    // M1-15 QA D1: a cold start whose Keystore read outlasts the restore
    // timeout showed login although the token was on disk.
    testWidgets('a slow secure-store read shows login, then home once the '
        'stored session arrives', (tester) async {
      harness.tokens.readGate = Completer<void>();
      await harness.pump(tester, signedIn: true);
      expect(find.byType(LoginScreen), findsOneWidget, reason: 'bounded');

      harness.tokens.readGate!.complete();
      await tester.pumpAndSettle();

      expect(find.byType(HomeScreen), findsOneWidget);
      expect(signOut(), findsOneWidget);
    });

    for (final location in [AuthRoutes.login, AuthRoutes.register]) {
      testWidgets('signed in: $location sends home', (tester) async {
        await harness.pump(tester, signedIn: true, location: location);

        expect(find.byType(HomeScreen), findsOneWidget);
      });
    }

    testWidgets('a login link never leaves the app (desde)', (tester) async {
      await harness.pump(tester);

      router().go('${AuthRoutes.login}?desde=https://evil.example');
      await tester.pumpAndSettle();
      await harness.session.start(sampleSession.tokens);
      await tester.pumpAndSettle();

      expect(find.byType(HomeScreen), findsOneWidget);
    });

    testWidgets('a rejected refresh ends the session and shows login', (
      tester,
    ) async {
      harness.rotation.error = const ApiUnauthorized();
      await harness.pump(tester, signedIn: true);

      await harness.session.refresh();
      await tester.pumpAndSettle();

      expect(find.byType(LoginScreen), findsOneWidget);
      expect(harness.tokens.token, isNull);
    });
  });

  group('logout', () {
    testWidgets('revokes on the server, wipes the store, shows login', (
      tester,
    ) async {
      await harness.pump(tester, signedIn: true);
      await harness.session.refresh(); // an access token in memory

      await tester.tap(signOut());
      await tester.pumpAndSettle();

      expect(harness.repository.loggedOut, ['refresh-rotated']);
      expect(harness.tokens.token, isNull);
      expect(harness.session.accessToken, isNull);
      expect(find.byType(LoginScreen), findsOneWidget);
    });

    testWidgets('after a cold start it gets an access token first', (
      tester,
    ) async {
      await harness.pump(tester, signedIn: true);

      await tester.tap(signOut());
      await tester.pumpAndSettle();

      expect(harness.rotation.received, [storedRefreshToken]);
      expect(harness.repository.loggedOut, ['refresh-rotated']);
    });

    for (final failure in <ApiFailure>[
      const ApiUnreachable(),
      const ApiUnauthorized(),
      const ApiUnexpected(),
    ]) {
      testWidgets('still wipes locally when the API fails '
          '(${failure.runtimeType})', (tester) async {
        harness.repository.logoutError = failure;
        await harness.pump(tester, signedIn: true);

        await tester.tap(signOut());
        await tester.pumpAndSettle();

        expect(harness.tokens.token, isNull);
        expect(harness.session.status, SessionStatus.signedOut);
        expect(find.byType(LoginScreen), findsOneWidget);
      });
    }

    testWidgets('still wipes locally when the refresh cannot happen', (
      tester,
    ) async {
      harness.rotation.error = const ApiUnreachable();
      await harness.pump(tester, signedIn: true);

      await tester.tap(signOut());
      await tester.pumpAndSettle();

      expect(harness.tokens.token, isNull);
      expect(find.byType(LoginScreen), findsOneWidget);
    });
  });

  group('module_auth off', () {
    testWidgets('home needs no session and offers no sign-out', (tester) async {
      await harness.pump(tester, enabled: false);

      expect(find.byType(HomeScreen), findsOneWidget);
      expect(signOut(), findsNothing);
    });
  });
}
