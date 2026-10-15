// Integration test against the Prism mock (`make mock`; see
// docs/runbooks/mock-server.md): the annual estimate and its breakdown
// (ANN-17, ANN-18) through the generated client and the app's own
// interceptor, validated by the frozen contract (0.11.0). Runs only when
// MOCK_API_URL is set: `make mobile-test-mock`.
@Tags(['mock'])
library;

import 'dart:io';

import 'package:control_contable/api/generated/models/annual_computed_section_kind.dart';
import 'package:control_contable/api/generated/models/disclaimer_code.dart';
import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/core/session/refresh_token_store.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/annual/data/annual_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../helpers/session_fakes.dart';

final String _mockUrl = Platform.environment['MOCK_API_URL'] ?? '';

const String _taxpayerId = '0199d4a2-8f00-7c44-a1b2-3c4d5e6f7a80';

const String _annualLegend =
    'Esto es un aproximado; tu contador te dará las cifras reales.';

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
    'the annual estimate against the contract mock',
    () {
      test('decodes with the annual legend, annual_estimate first', () async {
        final estimate = await container
            .read(annualRepositoryProvider)
            .getEstimate(taxpayerId: _taxpayerId, fiscalYear: 2026);

        expect(estimate.legend, _annualLegend);
        expect(estimate.disclaimers.first.code, DisclaimerCode.annualEstimate);
        expect(estimate.sections, isNotEmpty);
      });

      test('the breakdown decodes with its legend and lines', () async {
        final breakdown = await container
            .read(annualRepositoryProvider)
            .getBreakdown(
              taxpayerId: _taxpayerId,
              fiscalYear: 2026,
              section: AnnualComputedSectionKind.pfArt152,
            );

        expect(breakdown.legend, _annualLegend);
        expect(breakdown.disclaimers, isNotEmpty);
      });
    },
    skip: _mockUrl.isEmpty
        ? 'needs the Prism mock: make mock, then make mobile-test-mock'
        : false,
  );
}
