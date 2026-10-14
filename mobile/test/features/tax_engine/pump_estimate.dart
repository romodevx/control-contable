import 'dart:async';

import 'package:control_contable/api/generated/models/breakdown.dart';
import 'package:control_contable/api/generated/models/tax_input_value.dart';
import 'package:control_contable/api/generated/models/tax_inputs.dart';
import 'package:control_contable/api/generated/models/taxpayer_estimate.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/links/url_opener.dart';
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
import 'breakdown_fixtures.dart';
import 'estimate_fixtures.dart';
import 'tax_inputs_fixtures.dart';

/// One recorded estimate request.
typedef EstimateCall = ({String taxpayerId, String? period});

/// One recorded breakdown request.
typedef BreakdownCall = ({String taxpayerId, String period, String regime});

/// One recorded tax-inputs read.
typedef InputsCall = ({String taxpayerId, String regime, int fiscalYear});

/// One recorded tax-inputs save.
typedef SaveCall = ({
  String taxpayerId,
  String regime,
  int fiscalYear,
  List<TaxInputValue> values,
});

/// A [TaxEngineRepository] the test scripts. Each request answers with the
/// next scripted reply of its kind (the model, an `ApiFailure` to throw, or
/// a [Completer] still pending), else with its fallback; every call is
/// recorded.
class FakeTaxEngineRepository implements TaxEngineRepository {
  /// Estimate replies in order; then [fallback].
  final List<Object> replies = [];

  /// Every estimate request, in order.
  final List<EstimateCall> calls = [];

  /// The estimate when nothing is scripted: TX-612-TOTAL-01.
  Object fallback = estimateFrom(completeEstimateJson);

  /// Breakdown replies in order; then [breakdownFallback].
  final List<Object> breakdownReplies = [];

  /// Every breakdown request, in order.
  final List<BreakdownCall> breakdownCalls = [];

  /// The breakdown when nothing is scripted: 612, March 2026.
  Object breakdownFallback = breakdownFrom(breakdownJson());

  /// Tax-inputs read replies in order; then [inputsFallback].
  final List<Object> inputsReplies = [];

  /// Every tax-inputs read, in order.
  final List<InputsCall> inputsCalls = [];

  /// The tax inputs when nothing is scripted: 612, 2026, January paid.
  Object inputsFallback = taxInputsFrom(taxInputsJson());

  /// Save replies in order; then the saved values echoed back.
  final List<Object> saveReplies = [];

  /// Every save, in order.
  final List<SaveCall> saves = [];

  @override
  Future<TaxpayerEstimate> getEstimate({
    required String taxpayerId,
    String? period,
  }) {
    calls.add((taxpayerId: taxpayerId, period: period));
    return _answer(replies, fallback);
  }

  @override
  Future<Breakdown> getBreakdown({
    required String taxpayerId,
    required String period,
    required String regime,
  }) {
    breakdownCalls.add((
      taxpayerId: taxpayerId,
      period: period,
      regime: regime,
    ));
    return _answer(breakdownReplies, breakdownFallback);
  }

  @override
  Future<TaxInputs> getTaxInputs({
    required String taxpayerId,
    required String regime,
    required int fiscalYear,
  }) {
    inputsCalls.add((
      taxpayerId: taxpayerId,
      regime: regime,
      fiscalYear: fiscalYear,
    ));
    return _answer(inputsReplies, inputsFallback);
  }

  @override
  Future<TaxInputs> putTaxInputs({
    required String taxpayerId,
    required String regime,
    required int fiscalYear,
    required List<TaxInputValue> values,
  }) {
    saves.add((
      taxpayerId: taxpayerId,
      regime: regime,
      fiscalYear: fiscalYear,
      values: values,
    ));
    final stored = inputsFallback;
    final echo = stored is TaxInputs ? stored.copyWith(values: values) : stored;
    return _answer(saveReplies, echo);
  }

  static Future<T> _answer<T>(List<Object> replies, Object fallback) {
    final reply = replies.isEmpty ? fallback : replies.removeAt(0);
    return switch (reply) {
      final T value => Future.value(value),
      final Completer<T> pending => pending.future,
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

  /// Every web page the app asked to open.
  final List<Uri> openedUrls = [];

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
        urlOpenerProvider.overrideWithValue((url) async {
          openedUrls.add(url);
          return true;
        }),
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

  /// Pushes [location] on top of what is open.
  Future<void> push(WidgetTester tester, String location) async {
    unawaited(container.read(appRouterProvider).push<void>(location));
    await tester.pumpAndSettle();
  }

  /// The location on top (a pushed one included).
  String get location => container.read(appRouterProvider).state.uri.toString();
}
