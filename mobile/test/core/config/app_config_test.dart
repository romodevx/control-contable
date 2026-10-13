import 'package:control_contable/core/config/app_config.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('AppConfig.resolve', () {
    test('dev talks to the local API by default', () {
      final config = AppConfig.resolve(flavor: AppFlavor.dev);

      expect(config.flavor, AppFlavor.dev);
      expect(config.apiBaseUrl, Uri.parse('http://localhost:3000'));
    });

    test('dev talks to the Prism mock when asked to', () {
      final config = AppConfig.resolve(flavor: AppFlavor.dev, useMock: true);

      expect(config.apiBaseUrl, Uri.parse('http://localhost:3001'));
    });

    test('an explicit base URL wins (e.g. 10.0.2.2 on an emulator)', () {
      final config = AppConfig.resolve(
        flavor: AppFlavor.dev,
        apiBaseUrlOverride: 'http://10.0.2.2:3000',
        useMock: true,
      );

      expect(config.apiBaseUrl, Uri.parse('http://10.0.2.2:3000'));
    });

    for (final flavor in [AppFlavor.staging, AppFlavor.prod]) {
      test('${flavor.name} refuses to boot without a base URL', () {
        expect(
          () => AppConfig.resolve(flavor: flavor),
          throwsA(isA<AppConfigError>()),
        );
      });

      test('${flavor.name} refuses a plain-HTTP base URL', () {
        expect(
          () => AppConfig.resolve(
            flavor: flavor,
            apiBaseUrlOverride: 'http://api.example.com',
          ),
          throwsA(isA<AppConfigError>()),
        );
      });

      test('${flavor.name} ignores the mock switch', () {
        final config = AppConfig.resolve(
          flavor: flavor,
          apiBaseUrlOverride: 'https://api.example.com',
          useMock: true,
        );

        expect(config.apiBaseUrl, Uri.parse('https://api.example.com'));
      });
    }

    test('a malformed base URL is rejected', () {
      expect(
        () => AppConfig.resolve(
          flavor: AppFlavor.dev,
          apiBaseUrlOverride: 'localhost',
        ),
        throwsA(isA<AppConfigError>()),
      );
    });
  });
}
