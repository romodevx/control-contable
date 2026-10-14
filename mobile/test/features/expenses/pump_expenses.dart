import 'dart:async';

import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/expenses/data/expenses_repository.dart';
import 'package:control_contable/features/expenses/expenses_routes.dart';
import 'package:control_contable/features/taxpayers/application/client_id.dart';
import 'package:control_contable/features/taxpayers/data/active_taxpayer_store.dart';
import 'package:control_contable/features/taxpayers/data/taxpayers_repository.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/pump_app.dart';
import '../taxpayers/taxpayers_fakes.dart';
import 'expenses_fakes.dart';

/// "Now" in every expenses test: 2026-10-13 12:00 in Mexico City.
final DateTime expensesNow = DateTime.utc(2026, 10, 13, 18);

/// A persona física holding 612 and 606 since January 2026.
final Taxpayer anaTwoRegimes = taxpayer(
  regimes: [heldRegime('612'), heldRegime('606')],
);

/// A persona física holding only 612.
final Taxpayer anaOneRegime = taxpayer(regimes: [heldRegime('612')]);

/// A persona física in RESICO (626 PF).
final Taxpayer anaResico = taxpayer(regimes: [heldRegime('626')]);

/// A persona moral holding 601.
final Taxpayer moral601 = taxpayer(
  personType: PersonType.moral,
  displayName: 'Taller SA',
  regimes: [heldRegime('601')],
);

/// What an expenses widget test scripts and observes.
class ExpensesHarness {
  final FakeExpensesRepository expenses = FakeExpensesRepository();
  final FakeTaxpayersRepository taxpayers = FakeTaxpayersRepository();

  /// Client ids handed to each new form session, in order.
  final List<String> clientIds = [
    'bbbbbbbb-0000-4000-8000-000000000001',
    'bbbbbbbb-0000-4000-8000-000000000002',
    'bbbbbbbb-0000-4000-8000-000000000003',
  ];
  int _clientIdIndex = 0;

  late ProviderContainer container;

  /// Boots the app with taxpayers and expenses [enabled], the server
  /// holding [taxpayer] (612 + 606 by default), and opens [location].
  Future<void> pump(
    WidgetTester tester, {
    bool enabled = true,
    Taxpayer? taxpayer,
    String? location = ExpensesRoutes.list,
    Brightness brightness = Brightness.light,
  }) async {
    taxpayers.taxpayers.add(taxpayer ?? anaTwoRegimes);
    await pumpControlContableApp(
      tester,
      brightness: brightness,
      enabledModules: {AppModule.taxpayers, if (enabled) AppModule.expenses},
      overrides: [
        taxpayersRepositoryProvider.overrideWithValue(taxpayers),
        expensesRepositoryProvider.overrideWithValue(expenses),
        clientIdProvider.overrideWithValue(
          () => clientIds[_clientIdIndex++ % clientIds.length],
        ),
        clockProvider.overrideWithValue(() => expensesNow),
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

  /// The location on top.
  String get location => container.read(appRouterProvider).state.uri.toString();
}
