import 'dart:async';

import 'package:control_contable/api/generated/models/taxpayer_estimate.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/tax_engine/data/tax_engine_repository.dart';
import 'package:control_contable/features/tax_engine/tax_engine_routes.dart';
import 'package:control_contable/features/taxpayers/data/active_taxpayer_store.dart';
import 'package:control_contable/features/taxpayers/data/taxpayers_repository.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/pump_app.dart';
import '../taxpayers/taxpayers_fakes.dart';
import 'estimate_fixtures.dart';

/// One recorded estimate request.
typedef EstimateCall = ({String taxpayerId, String? period});

/// A [TaxEngineRepository] the test scripts. Each request answers with the
/// next scripted reply ([TaxpayerEstimate], an `ApiFailure` to throw, or a
/// [Completer] still pending), else with [fallback]; every call is recorded.
class FakeTaxEngineRepository implements TaxEngineRepository {
  /// Replies in order; the last one repeats once the queue is empty.
  final List<Object> replies = [];

  /// Every request, in order.
  final List<EstimateCall> calls = [];

  /// The reply when nothing is scripted: TX-612-TOTAL-01.
  Object fallback = estimateFrom(completeEstimateJson);

  @override
  Future<TaxpayerEstimate> getEstimate({
    required String taxpayerId,
    String? period,
  }) {
    calls.add((taxpayerId: taxpayerId, period: period));
    final reply = replies.isEmpty ? fallback : replies.removeAt(0);
    return switch (reply) {
      final TaxpayerEstimate estimate => Future.value(estimate),
      final Completer<TaxpayerEstimate> pending => pending.future,
      final Exception failure => Future.error(failure),
      _ => throw StateError('unexpected reply $reply'),
    };
  }
}

/// What an estimate widget test scripts and observes.
class EstimateHarness {
  /// The taxpayers server (one taxpayer holding 612).
  final FakeTaxpayersRepository taxpayers = FakeTaxpayersRepository();

  /// The estimates server.
  final FakeTaxEngineRepository repository = FakeTaxEngineRepository();

  /// The running app's providers.
  late ProviderContainer container;

  /// "Now": 2026-03-20 noon in Mexico City (TX-612-TOTAL-01's March).
  DateTime now = DateTime.utc(2026, 3, 20, 18);

  /// Boots the app with `module_tax_engine` [enabled] (and taxpayers) and
  /// opens [location] (the estimate by default) on top of home.
  Future<void> pump(
    WidgetTester tester, {
    bool enabled = true,
    String? location = TaxEngineRoutes.estimate,
    Brightness brightness = Brightness.light,
    Object? reply,
  }) async {
    if (reply != null) repository.fallback = reply;
    taxpayers.taxpayers.add(taxpayer(regimes: [heldRegime('612')]));
    await pumpControlContableApp(
      tester,
      brightness: brightness,
      enabledModules: {AppModule.taxpayers, if (enabled) AppModule.taxEngine},
      overrides: [
        taxpayersRepositoryProvider.overrideWithValue(taxpayers),
        taxEngineRepositoryProvider.overrideWithValue(repository),
        clockProvider.overrideWithValue(() => now),
        activeTaxpayerStoreProvider.overrideWithValue(
          InMemoryActiveTaxpayerStore(),
        ),
      ],
    );
    container = ProviderScope.containerOf(
      tester.element(find.byType(MaterialApp)),
    );
    if (location != null) {
      unawaited(container.read(appRouterProvider).push<void>(location));
      await tester.pumpAndSettle();
    }
  }

  /// The location on top (a pushed one included).
  String get location => container.read(appRouterProvider).state.uri.toString();
}
