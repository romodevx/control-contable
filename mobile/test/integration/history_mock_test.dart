// Integration test against the Prism mock (`make mock`; see
// docs/runbooks/mock-server.md): History (M8-05) through the generated
// client and the app's own interceptor, validated by the frozen contract
// (0.9.0). Runs only when MOCK_API_URL is set: `make mobile-test-mock`.
@Tags(['mock'])
library;

import 'dart:io';

import 'package:control_contable/api/generated/models/history_viewed_parameter_set.dart';
import 'package:control_contable/api/generated/models/history_viewed_request.dart';
import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/core/session/refresh_token_store.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/history/data/history_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../helpers/session_fakes.dart';

final String _mockUrl = Platform.environment['MOCK_API_URL'] ?? '';

const String _taxpayerId = '0199d4a2-8f00-7c44-a1b2-3c4d5e6f7a80';

final Object _skip = _mockUrl.isEmpty
    ? 'needs the Prism mock: make mock, then make mobile-test-mock'
    : false;

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

  test('History decodes the contract example, newest first', () async {
    final year = await container
        .read(historyRepositoryProvider)
        .getHistory(taxpayerId: _taxpayerId, fiscalYear: 2026);

    expect(year.months, isNotEmpty);
    final periods = year.months.map((m) => m.period).toList();
    expect(periods, [...periods]..sort((a, b) => b.compareTo(a)));
    expect(year.legend, startsWith('Esto es un aproximado.'));
  }, skip: _skip);

  test('marking a month viewed is accepted with the closed body', () async {
    await container
        .read(historyRepositoryProvider)
        .markViewed(
          taxpayerId: _taxpayerId,
          period: '2026-02',
          request: HistoryViewedRequest(
            recordsFingerprint: 'ab' * 32,
            figuresFingerprint: 'cd' * 32,
            engineVersion: '1.0.0',
            parameterSet: const HistoryViewedParameterSet(
              fiscalYear: 2026,
              version: 1,
            ),
          ),
        );
  }, skip: _skip);
}
