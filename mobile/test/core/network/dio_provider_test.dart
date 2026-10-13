import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/core/network/dio_provider.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('dio targets the configured API base URL', () {
    final container = ProviderContainer.test(
      overrides: [
        appConfigProvider.overrideWithValue(
          AppConfig.resolve(flavor: AppFlavor.dev, useMock: true),
        ),
      ],
    );

    final dio = container.read(dioProvider);

    expect(dio.options.baseUrl, 'http://localhost:3001');
    expect(dio.options.connectTimeout, isNotNull);
    expect(dio.options.receiveTimeout, isNotNull);
  });

  test('reading the config without an override fails loudly', () {
    final container = ProviderContainer.test();

    expect(() => container.read(appConfigProvider), throwsA(anything));
  });
}
