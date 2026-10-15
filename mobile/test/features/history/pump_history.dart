import 'dart:async';

import 'package:control_contable/api/generated/models/history_viewed_request.dart';
import 'package:control_contable/api/generated/models/history_year.dart';
import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/expenses/data/expenses_repository.dart';
import 'package:control_contable/features/history/data/history_repository.dart';
import 'package:control_contable/features/history/history_routes.dart';
import 'package:control_contable/features/incomes/data/incomes_repository.dart';
import 'package:control_contable/features/tax_engine/data/tax_engine_repository.dart';
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
import 'history_fixtures.dart';

/// One recorded history read.
typedef HistoryCall = ({String taxpayerId, int fiscalYear});

/// One recorded "seen" post.
typedef ViewedCall = ({
  String taxpayerId,
  String period,
  HistoryViewedRequest request,
});

/// A [HistoryRepository] the test scripts: each read answers with the next
/// scripted reply (a [HistoryYear], an exception, or a pending [Completer]),
/// else [fallback]; each "seen" post with the next of [viewedReplies]
/// (an exception or a pending [Completer]), else at once. All recorded.
class FakeHistoryRepository implements HistoryRepository {
  /// Read replies in order; then [fallback].
  final List<Object> replies = [];

  /// Every read, in order.
  final List<HistoryCall> calls = [];

  /// The read when nothing is scripted.
  Object fallback = historyFrom(historyYearJson());

  /// "Seen" replies in order; then success.
  final List<Object> viewedReplies = [];

  /// Every "seen" post, in order.
  final List<ViewedCall> viewed = [];

  @override
  Future<HistoryYear> getHistory({
    required String taxpayerId,
    required int fiscalYear,
  }) {
    calls.add((taxpayerId: taxpayerId, fiscalYear: fiscalYear));
    final reply = replies.isEmpty ? fallback : replies.removeAt(0);
    return switch (reply) {
      final HistoryYear value => Future.value(value),
      final Completer<HistoryYear> pending => pending.future,
      final Exception failure => Future.error(failure),
      _ => throw StateError('unexpected reply $reply'),
    };
  }

  @override
  Future<void> markViewed({
    required String taxpayerId,
    required String period,
    required HistoryViewedRequest request,
  }) {
    viewed.add((taxpayerId: taxpayerId, period: period, request: request));
    if (viewedReplies.isEmpty) return Future.value();
    return switch (viewedReplies.removeAt(0)) {
      final Completer<void> pending => pending.future,
      final Exception failure => Future.error(failure),
      final reply => throw StateError('unexpected reply $reply'),
    };
  }
}

/// The modules History needs, and the record lists it links to.
const Set<AppModule> historyModules = {
  AppModule.taxpayers,
  AppModule.taxEngine,
  AppModule.history,
  AppModule.incomes,
  AppModule.expenses,
};

/// What a history widget test scripts and observes.
class HistoryHarness {
  /// The taxpayers server.
  final FakeTaxpayersRepository taxpayers = FakeTaxpayersRepository();

  /// The history server.
  final FakeHistoryRepository repository = FakeHistoryRepository();

  /// The estimate server (behind each month).
  final FakeTaxEngineRepository estimates = FakeTaxEngineRepository();

  /// The incomes server (behind "Registrar ingreso" and "Ver ingresos").
  final FakeIncomesRepository incomes = FakeIncomesRepository();

  /// The expenses server (behind "Ver gastos").
  final FakeExpensesRepository expenses = FakeExpensesRepository();

  /// The running app's providers.
  late ProviderContainer container;

  /// "Now": 2026-03-20 noon in Mexico City.
  DateTime now = DateTime.utc(2026, 3, 20, 18);

  /// Boots the app with [modules] enabled, the server holding [held] (one
  /// taxpayer with 612 by default) and answering [reply], then opens
  /// [location] (History by default) on top of home.
  Future<void> pump(
    WidgetTester tester, {
    Set<AppModule> modules = historyModules,
    List<Taxpayer>? held,
    Object? reply,
    String? location = HistoryRoutes.list,
    Brightness brightness = Brightness.light,
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
        historyRepositoryProvider.overrideWithValue(repository),
        taxEngineRepositoryProvider.overrideWithValue(estimates),
        incomesRepositoryProvider.overrideWithValue(incomes),
        expensesRepositoryProvider.overrideWithValue(expenses),
        clockProvider.overrideWithValue(() => now),
        activeTaxpayerStoreProvider.overrideWithValue(
          InMemoryActiveTaxpayerStore(),
        ),
      ],
    );
    container = ProviderScope.containerOf(
      tester.element(find.byType(MaterialApp)),
    );
    if (location != null) await push(tester, location, settle: settle);
  }

  /// Pushes [location] on top of what is open.
  Future<void> push(
    WidgetTester tester,
    String location, {
    bool settle = true,
  }) async {
    unawaited(container.read(appRouterProvider).push<void>(location));
    if (settle) {
      await tester.pumpAndSettle();
    } else {
      await tester.pump();
      await tester.pump();
    }
  }

  /// Pops the page on top.
  Future<void> back(WidgetTester tester) async {
    container.read(appRouterProvider).pop();
    await tester.pumpAndSettle();
  }

  /// The location on top (a pushed one included).
  String get location => container.read(appRouterProvider).state.uri.toString();
}
