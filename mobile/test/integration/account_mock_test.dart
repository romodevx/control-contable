// Integration test against the Prism mock (`make mock`; see
// docs/runbooks/mock-server.md): profile read and rename, and account
// deletion through the generated client and the app's own interceptor,
// validated by the contract. Runs only when MOCK_API_URL is set:
// `make mobile-test-mock`.
@Tags(['mock'])
library;

import 'dart:io';

import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/core/session/refresh_token_store.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/auth/application/delete_account.dart';
import 'package:control_contable/features/auth/data/profile_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../helpers/session_fakes.dart';
import '../helpers/synthetic_password.dart';

final String _mockUrl = Platform.environment['MOCK_API_URL'] ?? '';

void main() {
  late ProviderContainer container;
  late InMemoryRefreshTokenStore store;

  setUp(() async {
    store = InMemoryRefreshTokenStore();
    container = ProviderContainer.test(
      overrides: [
        appConfigProvider.overrideWithValue(
          AppConfig.resolve(
            flavor: AppFlavor.dev,
            apiBaseUrlOverride: _mockUrl,
          ),
        ),
        refreshTokenStoreProvider.overrideWithValue(store),
      ],
    );
    await container
        .read(sessionProvider)
        .start(tokenPair(access: 'a', refresh: 'r'));
  });

  group(
    'account against the contract mock',
    () {
      test('the profile reads back with its email', () async {
        final user = await container.read(profileRepositoryProvider).getMe();

        expect(user.email, isNotEmpty);
      });

      test('a rename is accepted', () async {
        final user = await container
            .read(profileRepositoryProvider)
            .updateName('Ana Gómez de la Torre');

        expect(user.email, isNotEmpty);
      });

      test('deletion is accepted and forgets the session', () async {
        await container.read(deleteAccountProvider).call(syntheticPassword());

        expect(store.token, isNull);
        expect(container.read(sessionProvider).status, SessionStatus.signedOut);
      });
    },
    skip: _mockUrl.isEmpty ? 'set MOCK_API_URL (make mobile-test-mock)' : false,
  );
}
