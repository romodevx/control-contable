// Integration test against the Prism mock (`make mock`; see
// docs/runbooks/mock-server.md): the monthly estimate (M5-25) through the
// generated client and the app's own interceptor, validated by the frozen
// contract (0.6.0). Runs only when MOCK_API_URL is set:
// `make mobile-test-mock`.
@Tags(['mock'])
library;

import 'dart:io';

import 'package:control_contable/api/generated/models/disclaimer_code.dart';
import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/core/session/refresh_token_store.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/tax_engine/data/tax_engine_repository.dart';
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

  group(
    'the monthly estimate against the contract mock',
    () {
      for (final period in [null, '2026-03']) {
        test(
          'decodes with the legend and not_tax_advice (period: $period)',
          () async {
            final estimate = await container
                .read(taxEngineRepositoryProvider)
                .getEstimate(taxpayerId: _taxpayerId, period: period);

            expect(estimate.legend, startsWith('Esto es un aproximado.'));
            expect(
              estimate.disclaimers.first.code,
              DisclaimerCode.notTaxAdvice,
            );
            expect(estimate.totals.totalPayableRoundedCents, isNonNegative);
          },
        );
      }
    },
    skip: _mockUrl.isEmpty
        ? 'needs the Prism mock: make mock, then make mobile-test-mock'
        : false,
  );
}
