import 'package:control_contable/app/app.dart';
import 'package:control_contable/app/feature_modules.dart';
import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flags_repository.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/core/session/refresh_token_store.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences_platform_interface/in_memory_shared_preferences_async.dart';
import 'package:shared_preferences_platform_interface/shared_preferences_async_platform_interface.dart';

import 'fake_flags_repository.dart';
import 'session_fakes.dart';

/// The configuration every widget test boots with: the dev flavor, local API.
final AppConfig testConfig = AppConfig.resolve(flavor: AppFlavor.dev);

/// Boots the whole app the way `bootstrap` does, with [overrides] on top of
/// the test configuration, the platform in [brightness], and the server
/// answering `GET /v1/flags` with [enabledModules] (none by default), or
/// through [flagsRepository] when a test scripts several answers. The
/// module slices are the app's own unless a test passes its [features].
/// The session store is in memory ([refreshTokens], empty by default: signed
/// out), never the platform's; so are the plain preferences (empty: sign-out
/// reads the push installation marker, M9 security gate F1). [settle]
/// `false` pumps a few frames instead of settling, for a screen that shows
/// an endless spinner while it loads.
Future<void> pumpControlContableApp(
  WidgetTester tester, {
  Brightness brightness = Brightness.light,
  List<Override> overrides = const [],
  Set<AppModule> enabledModules = const {},
  FlagsRepository? flagsRepository,
  List<FeatureModule>? features,
  RefreshTokenStore? refreshTokens,
  bool settle = true,
}) async {
  tester.platformDispatcher.platformBrightnessTestValue = brightness;
  SharedPreferencesAsyncPlatform.instance =
      InMemorySharedPreferencesAsync.empty();
  addTearDown(tester.platformDispatcher.clearPlatformBrightnessTestValue);
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        appConfigProvider.overrideWithValue(testConfig),
        flagsRepositoryProvider.overrideWithValue(
          flagsRepository ?? FakeFlagsRepository(enabledModules),
        ),
        featureModulesProvider.overrideWithValue(features ?? appFeatureModules),
        ...appFeatureWiring,
        refreshTokenStoreProvider.overrideWithValue(
          refreshTokens ?? InMemoryRefreshTokenStore(),
        ),
        ...overrides,
      ],
      child: const ControlContableApp(),
    ),
  );
  if (settle) {
    await tester.pumpAndSettle();
  } else {
    for (var frame = 0; frame < _unsettledFrames; frame++) {
      await tester.pump();
    }
  }
}

/// Frames pumped without settling: enough for the flags, the session and
/// the first reads to resolve.
const int _unsettledFrames = 10;
