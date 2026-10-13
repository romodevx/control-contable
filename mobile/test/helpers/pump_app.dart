import 'package:control_contable/app/app.dart';
import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flags_repository.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';
import 'package:flutter_test/flutter_test.dart';

import 'fake_flags_repository.dart';

/// The configuration every widget test boots with: the dev flavor, local API.
final AppConfig testConfig = AppConfig.resolve(flavor: AppFlavor.dev);

/// Boots the whole app the way `bootstrap` does, with [overrides] on top of
/// the test configuration, the platform in [brightness], and the server
/// answering `GET /v1/flags` with [enabledModules] (none by default), or
/// through [flagsRepository] when a test scripts several answers.
Future<void> pumpControlContableApp(
  WidgetTester tester, {
  Brightness brightness = Brightness.light,
  List<Override> overrides = const [],
  Set<AppModule> enabledModules = const {},
  FlagsRepository? flagsRepository,
}) async {
  tester.platformDispatcher.platformBrightnessTestValue = brightness;
  addTearDown(tester.platformDispatcher.clearPlatformBrightnessTestValue);
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        appConfigProvider.overrideWithValue(testConfig),
        flagsRepositoryProvider.overrideWithValue(
          flagsRepository ?? FakeFlagsRepository(enabledModules),
        ),
        ...overrides,
      ],
      child: const ControlContableApp(),
    ),
  );
  await tester.pumpAndSettle();
}
