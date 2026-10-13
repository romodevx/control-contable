import 'package:control_contable/app/app.dart';
import 'package:control_contable/core/config/app_config.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';
import 'package:flutter_test/flutter_test.dart';

/// The configuration every widget test boots with: the dev flavor, local API.
final AppConfig testConfig = AppConfig.resolve(flavor: AppFlavor.dev);

/// Boots the whole app the way `bootstrap` does, with [overrides] on top of
/// the test configuration, and the platform in [brightness].
Future<void> pumpControlContableApp(
  WidgetTester tester, {
  Brightness brightness = Brightness.light,
  List<Override> overrides = const [],
}) async {
  tester.platformDispatcher.platformBrightnessTestValue = brightness;
  addTearDown(tester.platformDispatcher.clearPlatformBrightnessTestValue);
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        appConfigProvider.overrideWithValue(testConfig),
        ...overrides,
      ],
      child: const ControlContableApp(),
    ),
  );
  await tester.pumpAndSettle();
}
