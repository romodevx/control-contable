// Integration test against the Prism mock (`make mock`; see
// docs/runbooks/mock-server.md): the onboarding calls (regime catalog,
// create, RFC correction, tax inputs) through the generated client and the
// app's own interceptor, validated by the frozen contract (0.4.0). Runs
// only when MOCK_API_URL is set: `make mobile-test-mock`.
@Tags(['mock'])
library;

import 'dart:io';

import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/tax_input_value.dart';
import 'package:control_contable/api/generated/models/taxpayer_create.dart';
import 'package:control_contable/api/generated/models/taxpayer_regime_create.dart';
import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/core/session/refresh_token_store.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/taxpayers/application/client_id.dart';
import 'package:control_contable/features/taxpayers/data/taxpayers_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../helpers/session_fakes.dart';

final String _mockUrl = Platform.environment['MOCK_API_URL'] ?? '';

void main() {
  late ProviderContainer container;

  TaxpayersRepository repository() =>
      container.read(taxpayersRepositoryProvider);

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
    'taxpayer onboarding against the contract mock',
    () {
      for (final personType in [PersonType.fisica, PersonType.moral]) {
        test('the ${personType.name} catalog decodes', () async {
          final regimes = await repository().listRegimes(personType);

          expect(regimes, isNotEmpty);
        });
      }

      test('a create with regimes and options is accepted', () async {
        final created = await repository().createTaxpayer(
          TaxpayerCreate(
            id: newUuidV4(),
            personType: PersonType.fisica,
            rfc: 'GOMA800101AB6',
            displayName: 'Ana Gómez',
            regimes: const [
              TaxpayerRegimeCreate(
                regimeCode: '606',
                options: {'deduction_mode': 'blind_35'},
              ),
              TaxpayerRegimeCreate(regimeCode: '626'),
            ],
          ),
        );

        expect(created.rfcMasked, isNotEmpty);
      });

      test('an RFC correction is accepted', () async {
        final updated = await repository().correctRfc(
          taxpayerId: newUuidV4(),
          rfc: 'GOMA800101AB6',
        );

        expect(updated.rfcMasked, isNotEmpty);
      });

      test('tax inputs are accepted', () async {
        await repository().saveTaxInputs(
          taxpayerId: newUuidV4(),
          regimeCode: '601',
          fiscalYear: 2026,
          values: const [
            TaxInputValue(key: 'profit_coefficient', valueRate: '0.2345'),
          ],
        );
      });
    },
    skip: _mockUrl.isEmpty ? 'set MOCK_API_URL (make mobile-test-mock)' : false,
  );
}
