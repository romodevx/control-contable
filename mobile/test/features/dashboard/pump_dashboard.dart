import 'dart:async';

import 'package:control_contable/api/generated/models/dashboard.dart';
import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/dashboard/data/dashboard_repository.dart';
import 'package:control_contable/features/expenses/data/expenses_repository.dart';
import 'package:control_contable/features/incomes/data/incomes_repository.dart';
import 'package:control_contable/features/tax_engine/data/tax_engine_repository.dart';
import 'package:control_contable/features/taxpayers/application/client_id.dart';
import 'package:control_contable/features/taxpayers/data/active_taxpayer_store.dart';
import 'package:control_contable/features/taxpayers/data/taxpayers_repository.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/pump_app.dart';
import '../expenses/expenses_fakes.dart';
import '../incomes/incomes_fakes.dart';
import '../tax_engine/pump_estimate.dart';
import '../taxpayers/taxpayers_fakes.dart';
import 'dashboard_fixtures.dart';

/// A [DashboardRepository] the test scripts: each read answers with the
/// next scripted reply (a [Dashboard], an exception to throw, or a pending
/// [Completer]), else [fallback]. Every taxpayer id asked is recorded.
class FakeDashboardRepository implements DashboardRepository {
  /// Replies in order; then [fallback].
  final List<Object> replies = [];

  /// Every taxpayer id asked, in order.
  final List<String> calls = [];

  /// The reply when nothing is scripted.
  Object fallback = dashboardFrom(dashboardJson());

  @override
  Future<Dashboard> getDashboard({required String taxpayerId}) {
    calls.add(taxpayerId);
    final reply = replies.isEmpty ? fallback : replies.removeAt(0);
    return switch (reply) {
      final Dashboard value => Future.value(value),
      final Completer<Dashboard> pending => pending.future,
      final Exception failure => Future.error(failure),
      _ => throw StateError('unexpected reply $reply'),
    };
  }
}

/// Every module Home composes.
const Set<AppModule> homeModules = {
  AppModule.taxpayers,
  AppModule.dashboard,
  AppModule.taxEngine,
  AppModule.incomes,
  AppModule.expenses,
};

/// What a dashboard widget test scripts and observes.
class DashboardHarness {
  /// The taxpayers server.
  final FakeTaxpayersRepository taxpayers = FakeTaxpayersRepository();

  /// The dashboard server.
  final FakeDashboardRepository repository = FakeDashboardRepository();

  /// The estimate server (behind "Ver detalle").
  final FakeTaxEngineRepository estimates = FakeTaxEngineRepository();

  /// The incomes server (behind "Registrar ingreso").
  final FakeIncomesRepository incomes = FakeIncomesRepository();

  /// The expenses server (behind "Registrar gasto").
  final FakeExpensesRepository expenses = FakeExpensesRepository();

  /// The running app's providers.
  late ProviderContainer container;

  /// Boots the app on Home with [modules] enabled, the server holding
  /// [held] (one taxpayer with 612 by default) and answering [reply].
  Future<void> pump(
    WidgetTester tester, {
    Set<AppModule> modules = homeModules,
    List<Taxpayer>? held,
    Brightness brightness = Brightness.light,
    Object? reply,
    bool settle = true,
  }) async {
    if (reply != null) repository.fallback = reply;
    taxpayers.taxpayers.addAll(
      held ??
          [
            taxpayer(regimes: [heldRegime('612')]),
          ],
    );
    await pumpControlContableApp(
      tester,
      brightness: brightness,
      enabledModules: modules,
      settle: settle,
      overrides: [
        taxpayersRepositoryProvider.overrideWithValue(taxpayers),
        dashboardRepositoryProvider.overrideWithValue(repository),
        taxEngineRepositoryProvider.overrideWithValue(estimates),
        incomesRepositoryProvider.overrideWithValue(incomes),
        expensesRepositoryProvider.overrideWithValue(expenses),
        clientIdProvider.overrideWithValue(
          () => 'aaaaaaaa-0000-4000-8000-000000000001',
        ),
        clockProvider.overrideWithValue(() => DateTime.utc(2026, 3, 20, 18)),
        activeTaxpayerStoreProvider.overrideWithValue(
          InMemoryActiveTaxpayerStore(),
        ),
      ],
    );
    container = ProviderScope.containerOf(
      tester.element(find.byType(MaterialApp)),
    );
  }

  /// The location on top.
  String get location => container.read(appRouterProvider).state.uri.toString();

  /// Pops the page on top (back to Home).
  Future<void> back(WidgetTester tester) async {
    container.read(appRouterProvider).pop();
    await tester.pumpAndSettle();
  }
}
