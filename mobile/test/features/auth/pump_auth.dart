import 'package:control_contable/app/feature_modules.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/links/url_opener.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/auth/data/auth_repository.dart';
import 'package:control_contable/features/auth/data/profile_repository.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';

import '../../helpers/pump_app.dart';
import '../../helpers/session_fakes.dart';
import 'auth_fakes.dart';

/// What an auth widget test scripts and observes.
class AuthHarness {
  final FakeAuthRepository repository = FakeAuthRepository();
  final FakeProfileRepository profile = FakeProfileRepository();

  /// The secure store stand-in: what the app keeps of the session.
  final InMemoryRefreshTokenStore tokens = InMemoryRefreshTokenStore();
  final FakeTokenRotation rotation = FakeTokenRotation();
  final FakeClock clock = FakeClock();
  final List<Uri> opened = [];
  bool canOpen = true;

  /// The running app's providers.
  late ProviderContainer container;

  /// The app's session.
  Session get session => container.read(sessionProvider);

  /// The app's router.
  GoRouter get router => container.read(appRouterProvider);

  /// Boots the app with `module_auth` [enabled] (plus [alsoEnabled]),
  /// [signedIn] or not (a refresh token already stored), and opens
  /// [location]. The module slices are the app's own plus [extraFeatures].
  Future<void> pump(
    WidgetTester tester, {
    String? location,
    bool enabled = true,
    bool signedIn = false,
    Set<AppModule> alsoEnabled = const {},
    List<FeatureModule> extraFeatures = const [],
    Brightness brightness = Brightness.light,
    List<Override> overrides = const [],
  }) async {
    if (signedIn) tokens.token = storedRefreshToken;
    await pumpControlContableApp(
      tester,
      brightness: brightness,
      enabledModules: {if (enabled) AppModule.auth, ...alsoEnabled},
      refreshTokens: tokens,
      features: [...appFeatureModules, ...extraFeatures],
      overrides: [
        authRepositoryProvider.overrideWithValue(repository),
        profileRepositoryProvider.overrideWithValue(profile),
        tokenRotationProvider.overrideWithValue(rotation.call),
        clockProvider.overrideWithValue(clock.call),
        urlOpenerProvider.overrideWithValue((uri) async {
          opened.add(uri);
          return canOpen;
        }),
        ...overrides,
      ],
    );
    container = ProviderScope.containerOf(
      tester.element(find.byType(MaterialApp)),
    );
    if (location != null) {
      container.read(appRouterProvider).go(location);
      await tester.pumpAndSettle();
    }
  }

  /// Moves the clock and the test's fake time together.
  Future<void> elapse(WidgetTester tester, Duration by) async {
    clock.advance(by);
    await tester.pump(by);
  }
}

/// The text input under the visible [label] of an `AppTextField`.
Finder inputLabeled(String label) => find.descendant(
  of: find.ancestor(of: find.text(label), matching: find.byType(Column)).first,
  matching: find.byType(TextField),
);
