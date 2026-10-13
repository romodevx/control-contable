// Integration test against the Prism mock (`make mock`; see
// docs/runbooks/mock-server.md).
// Runs only when MOCK_API_URL is set: `make mobile-test-mock` from the repo
// root.
@Tags(['mock'])
library;

import 'dart:io';

import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flags_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

final String _mockUrl = Platform.environment['MOCK_API_URL'] ?? '';

void main() {
  test(
    'the generated client reads GET /v1/flags from the contract mock',
    () async {
      final container = ProviderContainer.test(
        overrides: [
          appConfigProvider.overrideWithValue(
            AppConfig.resolve(
              flavor: AppFlavor.dev,
              apiBaseUrlOverride: _mockUrl,
            ),
          ),
        ],
      );

      final enabled = await container
          .read(flagsRepositoryProvider)
          .fetchEnabled();

      // The contract's EnabledFlags example (static mock mode).
      expect(enabled, {
        AppModule.auth,
        AppModule.taxpayers,
        AppModule.incomes,
        AppModule.expenses,
        AppModule.taxEngine,
        AppModule.dashboard,
      });
    },
    skip: _mockUrl.isEmpty ? 'set MOCK_API_URL (make mobile-test-mock)' : false,
  );
}
