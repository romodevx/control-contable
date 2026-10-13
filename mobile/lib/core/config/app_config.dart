import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Build flavors. Each has its own entrypoint (`lib/main_<flavor>.dart`) and,
/// on Android, its own product flavor (`--flavor <name>`).
enum AppFlavor {
  /// Local development against the API on this machine.
  dev,

  /// The shared staging environment.
  staging,

  /// The store build.
  prod,
}

/// Raised at boot when the flavor and its compile-time defines do not make a
/// usable configuration. The app refuses to start rather than guess.
class AppConfigError extends Error {
  /// Creates the error with a [message] for the developer.
  new(this.message);

  /// What is wrong and how to fix it.
  final String message;

  @override
  String toString() => 'AppConfigError: $message';
}

/// Immutable runtime configuration, resolved once at boot.
@immutable
class AppConfig {
  const new _({required this.flavor, required this.apiBaseUrl});

  /// Resolves the configuration for [flavor].
  ///
  /// - dev: the local API (port 3000), or the Prism mock
  ///   (`make mock`, port 3001) when [useMock] is set.
  /// - staging / prod: [apiBaseUrlOverride] is required and must be HTTPS.
  ///   Their hosts are not baked into the source; CI passes
  ///   `--dart-define=API_BASE_URL=...`.
  ///
  /// A non-empty [apiBaseUrlOverride] always wins (an Android emulator reaches
  /// the host as `10.0.2.2`, not `localhost`).
  factory resolve({
    required AppFlavor flavor,
    String apiBaseUrlOverride = '',
    bool useMock = false,
  }) {
    final raw = apiBaseUrlOverride.isNotEmpty
        ? apiBaseUrlOverride
        : _defaultBaseUrl(flavor, useMock: useMock);
    return AppConfig._(flavor: flavor, apiBaseUrl: _parse(raw, flavor));
  }

  static const _localApi = 'http://localhost:3000';
  static const _localMock = 'http://localhost:3001';

  /// The flavor this build was started with.
  final AppFlavor flavor;

  /// Root of the REST API; the contract's paths (`/v1/...`) hang off it.
  final Uri apiBaseUrl;

  static String _defaultBaseUrl(AppFlavor flavor, {required bool useMock}) {
    if (flavor != AppFlavor.dev) {
      throw AppConfigError(
        '${flavor.name} needs --dart-define=API_BASE_URL=https://...',
      );
    }
    return useMock ? _localMock : _localApi;
  }

  static Uri _parse(String raw, AppFlavor flavor) {
    final uri = Uri.tryParse(raw);
    if (uri == null || !uri.hasScheme || uri.host.isEmpty) {
      throw AppConfigError('API_BASE_URL "$raw" is not an absolute URL');
    }
    if (flavor != AppFlavor.dev && uri.scheme != 'https') {
      throw AppConfigError('${flavor.name} must use HTTPS, got "$raw"');
    }
    return uri;
  }
}

/// The boot configuration. Scoped: `bootstrap` (and every test) overrides it;
/// reading it without an override is a programming error.
final appConfigProvider = Provider<AppConfig>(
  (ref) => throw UnimplementedError('appConfigProvider must be overridden'),
  name: 'appConfigProvider',
);
