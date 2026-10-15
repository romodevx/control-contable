import 'package:control_contable/app/app.dart';
import 'package:control_contable/app/feature_modules.dart';
import 'package:control_contable/app/font_licenses.dart';
import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/core/diagnostics/error_reporter.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/features/reminders/application/push_messaging.dart';
import 'package:control_contable/features/reminders/data/firebase_push_messaging.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Compile-time defines (`--dart-define=NAME=value`). See mobile/README.md.
const _apiBaseUrl = String.fromEnvironment('API_BASE_URL');
const _useMock = bool.fromEnvironment('USE_MOCK');

/// Shared entrypoint body: resolves the configuration for [flavor] (failing
/// fast when it is unusable), starts push when Firebase is configured
/// (never prompting: the permission is asked only when the user turns
/// reminders on), and starts the app with both.
Future<void> bootstrap(AppFlavor flavor) async {
  WidgetsFlutterBinding.ensureInitialized();
  registerFontLicenses();
  final config = AppConfig.resolve(
    flavor: flavor,
    // Equals the default only when no define is passed at build time.
    // ignore: avoid_redundant_argument_values
    apiBaseUrlOverride: _apiBaseUrl,
    // Equals the default only when no define is passed at build time.
    // ignore: avoid_redundant_argument_values
    useMock: _useMock,
  );
  final push = await startPushMessaging(reporter: const LogErrorReporter());
  runApp(
    ProviderScope(
      overrides: [
        appConfigProvider.overrideWithValue(config),
        featureModulesProvider.overrideWithValue(appFeatureModules),
        pushMessagingProvider.overrideWithValue(push),
        ...appFeatureWiring,
      ],
      child: const ControlContableApp(),
    ),
  );
}
