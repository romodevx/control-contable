// Integration test against the Prism mock (`make mock`; see
// docs/runbooks/mock-server.md): the expense calls of M4-09/M4-10 (list,
// categories, create, edit, delete, restore) through the generated client
// and the app's own interceptor, validated by the frozen contract (0.6.1).
// Runs only when MOCK_API_URL is set: `make mobile-test-mock`.
@Tags(['mock'])
library;

import 'dart:io';

import 'package:control_contable/api/generated/models/expense_create.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/api/generated/models/operation_kind.dart';
import 'package:control_contable/api/generated/models/payment_method.dart';
import 'package:control_contable/api/generated/models/supplier_kind.dart';
import 'package:control_contable/api/generated/models/supplier_regime.dart';
import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/core/session/refresh_token_store.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/expenses/data/expenses_repository.dart';
import 'package:control_contable/features/taxpayers/application/client_id.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../helpers/session_fakes.dart';

final String _mockUrl = Platform.environment['MOCK_API_URL'] ?? '';

const String _taxpayerId = '0192f0c4-7b1e-7c3a-9d2e-6f1a2b3c4d5e';

const String _expenseId = '0199d4a3-2b3c-7e4d-9f50-6b7c8d9e0f12';

void main() {
  late ProviderContainer container;

  ExpensesRepository repository() => container.read(expensesRepositoryProvider);

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
    'expenses against the contract mock',
    () {
      test('a month page decodes with its totals and legend', () async {
        final page = await repository().list(
          taxpayerId: _taxpayerId,
          period: '2026-10',
        );

        expect(page.totals.count, greaterThanOrEqualTo(0));
        expect(page.legend, isNotEmpty);
      });

      test('the categories are asked per taxpayer, regime and date', () async {
        final catalog = await repository().categories(
          taxpayerId: _taxpayerId,
          regime: '612',
          on: '2026-10-13',
        );

        expect(catalog.legend, isNotEmpty);
        expect(catalog.ivaRatesAllowed, isNotEmpty);
      });

      test(
        'a create with a client id and a YYYY-MM-DD date is accepted',
        () async {
          final created = await repository().create(
            taxpayerId: _taxpayerId,
            expense: ExpenseCreate(
              id: newUuidV4(),
              regimeCode: '612',
              categoryCode: 'software',
              amountCents: 50000,
              ivaRate: IvaRate.rate16,
              ivaCents: 8000,
              paymentMethod: PaymentMethod.cardDebit,
              paidOn: '2026-10-05',
              hasCfdi: true,
              description: 'Suscripción de diseño',
            ),
          );

          expect(created.amountCents, isPositive);
        },
      );

      test('the supplier withholding preview decodes '
          '(US-W6)', () async {
        final preview = await repository().previewWithholdings(
          taxpayerId: _taxpayerId,
          query: (
            regime: '601',
            on: '2026-06-15',
            amountCents: 4000000,
            ivaRate: IvaRate.rate16,
            ivaCents: 640000,
            categoryCode: 'servicios_profesionales',
            supplierKind: SupplierKind.pfWithActivity,
            supplierRegime: SupplierRegime.value612,
            operationKind: OperationKind.professionalServices,
            residentialRent: null,
          ),
        );

        expect(preview.withholding.netCents, isPositive);
        expect(preview.withholding.ruleIds, isNotEmpty);
      });

      test('an edit is a merge-patch that may clear the override', () async {
        final updated = await repository().update(
          taxpayerId: _taxpayerId,
          expenseId: _expenseId,
          patch: const {'amount_cents': 60000, 'deductible': null},
        );
        final stored = await repository().get(
          taxpayerId: _taxpayerId,
          expenseId: _expenseId,
        );

        expect(updated.id, isNotEmpty);
        expect(stored.id, isNotEmpty);
      });

      test('delete, then restore (Undo)', () async {
        await repository().delete(
          taxpayerId: _taxpayerId,
          expenseId: _expenseId,
        );
        final restored = await repository().restore(
          taxpayerId: _taxpayerId,
          expenseId: _expenseId,
        );

        expect(restored.id, isNotEmpty);
      });
    },
    skip: _mockUrl.isEmpty ? 'set MOCK_API_URL (make mobile-test-mock)' : false,
  );
}
