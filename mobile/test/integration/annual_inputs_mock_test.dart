// Integration test against the Prism mock (`make mock`; see
// docs/runbooks/mock-server.md): the annual inputs (ANN-19..21) through the
// generated client and the app's own interceptor, validated by the frozen
// contract (0.11.0). Runs only when MOCK_API_URL is set:
// `make mobile-test-mock`.
@Tags(['mock'])
library;

import 'dart:io';

import 'package:control_contable/api/generated/models/annual_inputs_update.dart';
import 'package:control_contable/api/generated/models/filed_payment_entry.dart';
import 'package:control_contable/api/generated/models/filed_payments_update.dart';
import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/core/session/refresh_token_store.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/annual/data/annual_inputs_repository.dart';
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

  AnnualInputsRepository repository() =>
      container.read(annualInputsRepositoryProvider);

  group(
    'the annual inputs against the contract mock',
    () {
      test('personal deductions and constancias decode', () async {
        final deductions = await repository().listDeductions(
          taxpayerId: _taxpayerId,
          fiscalYear: 2026,
        );
        final statements = await repository().listStatements(
          taxpayerId: _taxpayerId,
          fiscalYear: 2026,
        );

        expect(deductions.fiscalYear, 2026);
        expect(statements.totals.accumulableCents, isNonNegative);
      });

      test('the annual facts decode and replace', () async {
        final facts = await repository().getFacts(
          taxpayerId: _taxpayerId,
          fiscalYear: 2026,
        );
        final saved = await repository().putFacts(
          taxpayerId: _taxpayerId,
          fiscalYear: 2026,
          facts: const AnnualInputsUpdate(
            priorYearAccumulableIncomeCents: null,
            otherExemptIncomeCents: 0,
            toldEmployerWillFile: false,
          ),
        );

        expect(facts.otherExemptIncomeCents, isNonNegative);
        expect(saved.fiscalYear, isPositive);
      });

      test('the filed payments decode and replace', () async {
        final payments = await repository().getFiledPayments(
          taxpayerId: _taxpayerId,
          fiscalYear: 2026,
          regime: '612',
        );
        final saved = await repository().putFiledPayments(
          taxpayerId: _taxpayerId,
          fiscalYear: 2026,
          regime: '612',
          payments: const FiledPaymentsUpdate(
            months: [FiledPaymentEntry(month: 1, amountCents: 100000)],
          ),
        );

        expect(payments.months, isNotEmpty);
        expect(saved.months, isNotEmpty);
      });
    },
    skip: _mockUrl.isEmpty
        ? 'needs the Prism mock: make mock, then make mobile-test-mock'
        : false,
  );
}
