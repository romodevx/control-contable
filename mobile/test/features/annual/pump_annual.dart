import 'dart:async';

import 'package:control_contable/api/generated/models/annual_breakdown.dart';
import 'package:control_contable/api/generated/models/annual_computed_section_kind.dart';
import 'package:control_contable/api/generated/models/annual_estimate.dart';
import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/core/access/read_only_access.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/links/url_opener.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/annual/annual_routes.dart';
import 'package:control_contable/features/annual/data/annual_repository.dart';
import 'package:control_contable/features/taxpayers/data/active_taxpayer_store.dart';
import 'package:control_contable/features/taxpayers/data/taxpayers_repository.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/pump_app.dart';
import '../taxpayers/taxpayers_fakes.dart';
import 'annual_fixtures.dart';

/// One recorded annual estimate request.
typedef AnnualCall = ({String taxpayerId, int fiscalYear});

/// One recorded annual breakdown request.
typedef AnnualBreakdownCall = ({
  String taxpayerId,
  int fiscalYear,
  AnnualComputedSectionKind section,
});

/// An [AnnualRepository] the test scripts: each request answers with the
/// next scripted reply of its kind (the model, an exception to throw, or a
/// [Completer] still pending), else with its fallback. Every call is
/// recorded.
class FakeAnnualRepository implements AnnualRepository {
  /// Estimate replies in order; then [fallback].
  final List<Object> replies = [];

  /// Every estimate request, in order.
  final List<AnnualCall> calls = [];

  /// The estimate when nothing is scripted: AN-01.
  Object fallback = annualEstimateFrom(annualEstimateJson());

  /// Breakdown replies in order; then [breakdownFallback].
  final List<Object> breakdownReplies = [];

  /// Every breakdown request, in order.
  final List<AnnualBreakdownCall> breakdownCalls = [];

  /// The breakdown when nothing is scripted: AN-01's `pf_art152`.
  Object breakdownFallback = annualBreakdownFrom(annualBreakdownJson());

  /// What the input counters answer (or an exception to throw).
  Object counts = (personalDeductions: 3, salaryStatements: 1);

  @override
  Future<AnnualEstimate> getEstimate({
    required String taxpayerId,
    required int fiscalYear,
  }) {
    calls.add((taxpayerId: taxpayerId, fiscalYear: fiscalYear));
    return _answer(replies, fallback);
  }

  @override
  Future<AnnualBreakdown> getBreakdown({
    required String taxpayerId,
    required int fiscalYear,
    required AnnualComputedSectionKind section,
  }) {
    breakdownCalls.add((
      taxpayerId: taxpayerId,
      fiscalYear: fiscalYear,
      section: section,
    ));
    return _answer(breakdownReplies, breakdownFallback);
  }

  @override
  Future<AnnualInputCounts> getInputCounts({
    required String taxpayerId,
    required int fiscalYear,
  }) => _answer(const [], counts);

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

/// What an annual widget test scripts and observes.
class AnnualHarness {
  /// The taxpayers server (one persona física holding 612).
  final FakeTaxpayersRepository taxpayers = FakeTaxpayersRepository();

  /// The annual server.
  final FakeAnnualRepository repository = FakeAnnualRepository();

  /// The running app's providers.
  late ProviderContainer container;

  /// "Now": 2026-10-14 noon in Mexico City (year 2026 in progress).
  DateTime now = DateTime.utc(2026, 10, 14, 18);

  /// Whether the session already knows the user is read-only.
  bool readOnly = false;

  /// The active taxpayer's person type.
  PersonType personType = PersonType.fisica;

  /// Every web page the app asked to open.
  final List<Uri> openedUrls = [];

  /// Boots the app with `module_annual` [enabled] (with its prerequisites)
  /// and opens [location] (the annual estimate by default) on top of home.
  Future<void> pump(
    WidgetTester tester, {
    bool enabled = true,
    String? location = AnnualRoutes.estimate,
    Brightness brightness = Brightness.light,
    Object? reply,
  }) async {
    if (reply != null) repository.fallback = reply;
    taxpayers.taxpayers.add(
      taxpayer(personType: personType, regimes: [heldRegime('612')]),
    );
    await pumpControlContableApp(
      tester,
      brightness: brightness,
      enabledModules: {
        AppModule.taxpayers,
        AppModule.taxEngine,
        if (enabled) AppModule.annual,
      },
      overrides: [
        taxpayersRepositoryProvider.overrideWithValue(taxpayers),
        annualRepositoryProvider.overrideWithValue(repository),
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
    if (readOnly) {
      container.read(readOnlyAccessProvider.notifier).markReadOnly();
    }
    if (location != null) {
      unawaited(container.read(appRouterProvider).push<void>(location));
      await tester.pumpAndSettle();
    }
  }

  /// The location on top (a pushed one included).
  String get location => container.read(appRouterProvider).state.uri.toString();
}
