// Integration test against the Prism mock (`make mock`; see
// docs/runbooks/mock-server.md): the Home dashboard (M6-04) through the
// generated client and the app's own interceptor, validated by the frozen
// contract (0.8.0). Runs only when MOCK_API_URL is set:
// `make mobile-test-mock`.
@Tags(['mock'])
library;

import 'dart:io';

import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/core/session/refresh_token_store.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/dashboard/data/dashboard_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../helpers/session_fakes.dart';

final String _mockUrl = Platform.environment['MOCK_API_URL'] ?? '';

const String _taxpayerId = '0199d4a2-8f00-7c44-a1b2-3c4d5e6f7a80';

void main() {
  late ProviderContainer container;

  setUp(() async {
    container = ProviderContainer.test(
      overrides: [
        appConfigProvider.overrideWithValue(
          AppConfig.resolve(
            flavor: AppFlavor.dev,
            apiBaseUrlOverride: _mockUrl,
          ),
        ),
        refreshTokenStoreProvider.overrideWithValue(
          InMemoryRefreshTokenStore(),
        ),
      ],
    );
    await container
        .read(sessionProvider)
        .start(tokenPair(access: 'a', refresh: 'r'));
  });

  test(
    'Home decodes the contract example with its trend and access',
    () async {
      final dashboard = await container
          .read(dashboardRepositoryProvider)
          .getDashboard(taxpayerId: _taxpayerId);

      expect(dashboard.trend, isNotEmpty);
      expect(dashboard.trend.last.period, dashboard.period);
      if (dashboard.estimate != null) {
        expect(dashboard.legend, startsWith('Esto es un aproximado.'));
      }
    },
    skip: _mockUrl.isEmpty
        ? 'needs the Prism mock: make mock, then make mobile-test-mock'
        : false,
  );
}
