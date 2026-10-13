import 'dart:async';

import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flags_repository.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/core/router/boot_splash_screen.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/core/router/not_found_screen.dart';
import 'package:control_contable/features/home/presentation/home_screen.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';

import '../../helpers/pump_app.dart';

GoRouter _routerOf(WidgetTester tester) {
  final container = ProviderScope.containerOf(
    tester.element(find.byType(MaterialApp)),
  );
  return container.read(appRouterProvider);
}

void main() {
  testWidgets('the initial location is the home route', (tester) async {
    await pumpControlContableApp(tester);

    final router = _routerOf(tester);
    expect(router.routerDelegate.currentConfiguration.uri.path, AppRoutes.home);
    expect(find.byType(HomeScreen), findsOneWidget);
  });

  testWidgets('an unknown location shows the not-found screen', (tester) async {
    await pumpControlContableApp(tester);

    _routerOf(tester).go('/no-existe');
    await tester.pumpAndSettle();

    expect(find.byType(NotFoundScreen), findsOneWidget);
    expect(find.text('Página no encontrada'), findsOneWidget);

    await tester.tap(find.text('Ir al inicio'));
    await tester.pumpAndSettle();
    expect(find.byType(HomeScreen), findsOneWidget);
  });

  group('module routes follow the flags', () {
    const location = '/ingresos-prueba';
    final incomesFeature = FeatureModule(
      module: AppModule.incomes,
      routes: [
        GoRoute(
          path: location,
          builder: (context, state) => const Text('pantalla de ingresos'),
        ),
      ],
    );

    Future<void> visit(WidgetTester tester, Set<AppModule> enabled) async {
      await pumpControlContableApp(
        tester,
        enabledModules: enabled,
        overrides: [
          featureModulesProvider.overrideWithValue([incomesFeature]),
        ],
      );
      _routerOf(tester).go(location);
      await tester.pumpAndSettle();
    }

    testWidgets('a disabled module registers no routes (not found)', (
      tester,
    ) async {
      await visit(tester, const {});

      expect(find.byType(NotFoundScreen), findsOneWidget);
      expect(find.text('pantalla de ingresos'), findsNothing);
    });

    testWidgets('an enabled module registers its routes', (tester) async {
      await visit(tester, const {AppModule.incomes});

      expect(find.text('pantalla de ingresos'), findsOneWidget);
    });
  });

  group('a deep link before the flags load (F-01)', () {
    const location = '/ingresos-prueba';
    final features = [
      FeatureModule(
        module: AppModule.incomes,
        routes: [
          GoRoute(
            path: location,
            builder: (context, state) => const Text('pantalla de ingresos'),
          ),
        ],
      ),
    ];

    Future<Completer<Set<AppModule>>> deepLinkWhileLoading(
      WidgetTester tester,
    ) async {
      final pending = Completer<Set<AppModule>>();
      await pumpControlContableApp(
        tester,
        flagsRepository: _PendingFlagsRepository(pending.future),
        overrides: [featureModulesProvider.overrideWithValue(features)],
      );
      _routerOf(tester).go('$location?mes=2026-10');
      await tester.pump();
      return pending;
    }

    testWidgets('waits on the splash instead of not found', (tester) async {
      final pending = await deepLinkWhileLoading(tester);

      expect(find.byType(BootSplashScreen), findsOneWidget);
      expect(find.byType(NotFoundScreen), findsNothing);

      pending.complete(const {});
      await tester.pumpAndSettle();
    });

    testWidgets('lands on the module once its flag arrives ON', (tester) async {
      final pending = await deepLinkWhileLoading(tester);

      pending.complete({AppModule.incomes});
      await tester.pumpAndSettle();

      expect(find.text('pantalla de ingresos'), findsOneWidget);
      final uri = _routerOf(tester).routerDelegate.currentConfiguration.uri;
      expect(uri.toString(), '$location?mes=2026-10');
    });

    testWidgets('shows not found once its flag arrives OFF', (tester) async {
      final pending = await deepLinkWhileLoading(tester);

      pending.complete(const {});
      await tester.pumpAndSettle();

      expect(find.byType(NotFoundScreen), findsOneWidget);
    });

    testWidgets('fails closed to not found when the flags never arrive', (
      tester,
    ) async {
      await deepLinkWhileLoading(tester);

      await tester.pump(flagsBootTimeout);
      await tester.pumpAndSettle();

      expect(find.byType(NotFoundScreen), findsOneWidget);
      expect(find.text('pantalla de ingresos'), findsNothing);
    });

    for (final from in ['//evil.example/x', 'https://evil.example', '']) {
      testWidgets('the splash never leaves the app (desde="$from")', (
        tester,
      ) async {
        await pumpControlContableApp(tester);

        final splash = Uri(
          path: AppRoutes.splash,
          queryParameters: {AppRoutes.splashFrom: from},
        );
        _routerOf(tester).go(splash.toString());
        await tester.pumpAndSettle();

        expect(find.byType(HomeScreen), findsOneWidget);
      });
    }

    testWidgets('a direct visit to the splash also fails closed in time', (
      tester,
    ) async {
      await pumpControlContableApp(
        tester,
        flagsRepository: _PendingFlagsRepository(
          Completer<Set<AppModule>>().future,
        ),
        overrides: [featureModulesProvider.overrideWithValue(features)],
      );
      _routerOf(tester).go(
        Uri(
          path: AppRoutes.splash,
          queryParameters: {AppRoutes.splashFrom: location},
        ).toString(),
      );
      await tester.pump();
      expect(find.byType(BootSplashScreen), findsOneWidget);

      await tester.pump(flagsBootTimeout);
      await tester.pumpAndSettle();

      expect(find.byType(NotFoundScreen), findsOneWidget);
    });

    testWidgets('an unowned location does not wait for the flags', (
      tester,
    ) async {
      final pending = await deepLinkWhileLoading(tester);
      _routerOf(tester).go(AppRoutes.home);
      await tester.pump();

      expect(find.byType(HomeScreen), findsOneWidget);

      pending.complete(const {});
      await tester.pumpAndSettle();
    });
  });

  group('a flag change on resume', () {
    const incomesLocation = '/ingresos-prueba';
    final features = [
      FeatureModule(
        module: AppModule.incomes,
        routes: [
          GoRoute(
            path: incomesLocation,
            builder: (context, state) => const Text('pantalla de ingresos'),
          ),
        ],
      ),
      FeatureModule(
        module: AppModule.expenses,
        routes: [
          GoRoute(
            path: '/gastos-prueba',
            builder: (context, state) => const Text('pantalla de gastos'),
          ),
        ],
      ),
    ];

    Future<void> openIncomesThenResume(
      WidgetTester tester,
      Set<AppModule> afterResume,
    ) async {
      await pumpControlContableApp(
        tester,
        flagsRepository: _SequenceFlagsRepository([
          {AppModule.incomes},
          afterResume,
        ]),
        overrides: [featureModulesProvider.overrideWithValue(features)],
      );
      _routerOf(tester).go(incomesLocation);
      await tester.pumpAndSettle();
      expect(find.text('pantalla de ingresos'), findsOneWidget);

      tester.binding
        ..handleAppLifecycleStateChanged(AppLifecycleState.inactive)
        ..handleAppLifecycleStateChanged(AppLifecycleState.resumed);
      await tester.pumpAndSettle();
    }

    testWidgets('to another module keeps the user where they were', (
      tester,
    ) async {
      await openIncomesThenResume(tester, {
        AppModule.incomes,
        AppModule.expenses,
      });

      expect(find.text('pantalla de ingresos'), findsOneWidget);
    });

    testWidgets('turning off the open module shows not found', (tester) async {
      await openIncomesThenResume(tester, const {});

      expect(find.text('pantalla de ingresos'), findsNothing);
      expect(find.byType(NotFoundScreen), findsOneWidget);
    });
  });
}

/// Answers each fetch with the next set in [_answers].
class _SequenceFlagsRepository implements FlagsRepository {
  new(this._answers);

  final List<Set<AppModule>> _answers;
  int _calls = 0;

  @override
  Future<Set<AppModule>> fetchEnabled() async => _answers[_calls++];
}

/// Answers every fetch with [_answer], which a test completes when it wants.
class _PendingFlagsRepository implements FlagsRepository {
  new(this._answer);

  final Future<Set<AppModule>> _answer;

  @override
  Future<Set<AppModule>> fetchEnabled() => _answer;
}
