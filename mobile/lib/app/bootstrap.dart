import 'package:control_contable/app/app.dart';
import 'package:control_contable/app/feature_modules.dart';
import 'package:control_contable/app/font_licenses.dart';
import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Compile-time defines (`--dart-define=NAME=value`). See mobile/README.md.
const _apiBaseUrl = String.fromEnvironment('API_BASE_URL');
const _useMock = bool.fromEnvironment('USE_MOCK');

/// Shared entrypoint body: resolves the configuration for [flavor] (failing
/// fast when it is unusable) and starts the app with it.
void bootstrap(AppFlavor flavor) {
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
  runApp(
    ProviderScope(
      overrides: [
        appConfigProvider.overrideWithValue(config),
        featureModulesProvider.overrideWithValue(appFeatureModules),
        ...appFeatureWiring,
      ],
      child: const ControlContableApp(),
    ),
  );
}
