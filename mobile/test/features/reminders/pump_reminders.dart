import 'dart:async';

import 'package:control_contable/api/generated/models/device_platform.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/auth/data/auth_repository.dart';
import 'package:control_contable/features/reminders/application/push_device.dart';
import 'package:control_contable/features/reminders/application/push_messaging.dart';
import 'package:control_contable/features/reminders/data/installation_store.dart';
import 'package:control_contable/features/reminders/data/reminders_repository.dart';
import 'package:control_contable/features/reminders/presentation/system_settings_opener.dart';
import 'package:control_contable/features/reminders/reminders_routes.dart';
import 'package:control_contable/features/tax_engine/data/tax_engine_repository.dart';
import 'package:control_contable/features/taxpayers/data/active_taxpayer_store.dart';
import 'package:control_contable/features/taxpayers/data/taxpayers_repository.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/pump_app.dart';
import '../../helpers/session_fakes.dart';
import '../auth/auth_fakes.dart';
import '../tax_engine/pump_estimate.dart';
import '../taxpayers/taxpayers_fakes.dart';
import 'reminders_fakes.dart';

/// The installation id every reminders widget test runs with.
const String testInstallationId = '01920000-0000-7000-8000-0000000000aa';

/// The modules the reminders screen needs, the amounts toggle included.
const Set<AppModule> remindersModules = {
  AppModule.taxpayers,
  AppModule.taxEngine,
  AppModule.reminders,
};

/// What a reminders widget test scripts and observes.
class RemindersHarness {
  /// The reminders server.
  final FakeRemindersRepository repository = FakeRemindersRepository();

  /// The push provider (Firebase stand-in).
  final FakePushMessaging messaging = FakePushMessaging();

  /// This installation's identity on the device.
  final InMemoryInstallationStore store = InMemoryInstallationStore(
    testInstallationId,
  );

  /// The taxpayers server (Home's switcher reads it).
  final FakeTaxpayersRepository taxpayers = FakeTaxpayersRepository();

  /// The auth server (sign-out).
  final FakeAuthRepository auth = FakeAuthRepository();

  /// The session store; holds a token when the test boots signed in.
  final InMemoryRefreshTokenStore tokens = InMemoryRefreshTokenStore();

  /// How many times "Abrir ajustes del sistema" opened them.
  int settingsOpened = 0;

  /// "Now": 2026-05-01 noon in Mexico City.
  DateTime now = DateTime.utc(2026, 5, 1, 18);

  /// The running app's providers.
  late ProviderContainer container;

  /// Boots the app with [modules], [signedIn] or not, then opens
  /// [location] (the reminders screen by default) on top of Home.
  Future<void> pump(
    WidgetTester tester, {
    Set<AppModule> modules = remindersModules,
    String? location = RemindersRoutes.settings,
    bool signedIn = false,
  }) async {
    if (signedIn) tokens.token = 'refresh-stored';
    // One taxpayer, so Home never opens the onboarding over the test.
    taxpayers.taxpayers.add(taxpayer(regimes: [heldRegime('612')]));
    await pumpControlContableApp(
      tester,
      enabledModules: modules,
      refreshTokens: tokens,
      overrides: [
        remindersRepositoryProvider.overrideWithValue(repository),
        pushMessagingProvider.overrideWithValue(messaging),
        installationStoreProvider.overrideWithValue(store),
        devicePlatformProvider.overrideWithValue(DevicePlatform.android),
        systemSettingsOpenerProvider.overrideWithValue(
          () async => settingsOpened++,
        ),
        taxpayersRepositoryProvider.overrideWithValue(taxpayers),
        taxEngineRepositoryProvider.overrideWithValue(
          FakeTaxEngineRepository(),
        ),
        authRepositoryProvider.overrideWithValue(auth),
        tokenRotationProvider.overrideWithValue(FakeTokenRotation().call),
        activeTaxpayerStoreProvider.overrideWithValue(
          InMemoryActiveTaxpayerStore(),
        ),
        clockProvider.overrideWithValue(() => now),
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
