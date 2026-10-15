// Integration test against the Prism mock (`make mock`; see
// docs/runbooks/mock-server.md): the income calls of M3-08/M3-09 (list,
// defaults, create, delete, restore) through the generated client and the
// app's own interceptor, validated by the frozen contract (0.6.1). Runs
// only when MOCK_API_URL is set: `make mobile-test-mock`.
@Tags(['mock'])
library;

import 'dart:io';

import 'package:control_contable/api/generated/models/counterparty_kind.dart';
import 'package:control_contable/api/generated/models/income_create.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/api/generated/models/operation_kind.dart';
import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/core/session/refresh_token_store.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/incomes/data/incomes_repository.dart';
import 'package:control_contable/features/taxpayers/application/client_id.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../helpers/session_fakes.dart';

final String _mockUrl = Platform.environment['MOCK_API_URL'] ?? '';

const String _taxpayerId = '0192f0c4-7b1e-7c3a-9d2e-6f1a2b3c4d5e';

void main() {
  late ProviderContainer container;

  IncomesRepository repository() => container.read(incomesRepositoryProvider);

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
    'incomes against the contract mock',
    () {
      test('a month page decodes with its totals', () async {
        final page = await repository().list(
          taxpayerId: _taxpayerId,
          period: '2026-10',
        );

        expect(page.totals.count, greaterThanOrEqualTo(0));
      });

      test('the defaults are asked with a calendar date', () async {
        final defaults = await repository().defaults(
          taxpayerId: _taxpayerId,
          regime: '612',
          on: '2026-10-13',
        );

        expect(defaults.ivaRatesAllowed, isNotEmpty);
      });

      test(
        'a create with a client id and a YYYY-MM-DD date is accepted',
        () async {
          final created = await repository().create(
            taxpayerId: _taxpayerId,
            income: IncomeCreate(
              id: newUuidV4(),
              regimeCode: '612',
              counterpartyKind: CounterpartyKind.pm,
              operationKind: OperationKind.professionalServices,
              amountCents: 1000000,
              ivaRate: IvaRate.rate16,
              ivaCents: 160000,
              collectedOn: '2026-10-05',
              description: 'Diseño de logotipo',
            ),
          );

          expect(created.amountCents, isPositive);
        },
      );

      test("the withholding preview of the owner's case decodes "
          '(US-W1)', () async {
        final preview = await repository().previewWithholdings(
          taxpayerId: _taxpayerId,
          query: (
            regime: '626',
            on: '2026-07-10',
            amountCents: 10000000,
            ivaRate: IvaRate.rate16,
            ivaCents: 1600000,
            counterpartyKind: CounterpartyKind.pm,
            operationKind: OperationKind.professionalServices,
            residentialRent: null,
            viaPlatform: null,
            platformActivity: null,
          ),
        );

        expect(preview.withholding.netCents, isPositive);
        expect(preview.withholding.ruleIds, isNotEmpty);
      });

      test('an edit is a merge-patch with an explicit null, and the '
          'income reads back', () async {
        const incomeId = '0192f0c4-0000-7000-8000-000000000001';

        final updated = await repository().update(
          taxpayerId: _taxpayerId,
          incomeId: incomeId,
          patch: const {'amount_cents': 1200000, 'description': null},
        );
        final stored = await repository().get(
          taxpayerId: _taxpayerId,
          incomeId: incomeId,
        );

        expect(updated.id, isNotEmpty);
        expect(stored.id, isNotEmpty);
      });

      test('delete, then restore (Undo)', () async {
        const incomeId = '0192f0c4-0000-7000-8000-000000000001';

        await repository().delete(taxpayerId: _taxpayerId, incomeId: incomeId);
        final restored = await repository().restore(
          taxpayerId: _taxpayerId,
          incomeId: incomeId,
        );

        expect(restored.id, isNotEmpty);
      });
    },
    skip: _mockUrl.isEmpty ? 'set MOCK_API_URL (make mobile-test-mock)' : false,
  );
}
