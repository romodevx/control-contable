// Integration test against the Prism mock (`make mock`; see
// docs/runbooks/mock-server.md): login, refresh and logout through the
// generated client and the app's own interceptor, validated by the
// contract. Runs only when MOCK_API_URL is set: `make mobile-test-mock`.
@Tags(['mock'])
library;

import 'dart:io';

import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/core/session/refresh_token_store.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/auth/application/sign_out.dart';
import 'package:control_contable/features/auth/data/auth_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../helpers/session_fakes.dart';
import '../helpers/synthetic_password.dart';

final String _mockUrl = Platform.environment['MOCK_API_URL'] ?? '';

void main() {
  late ProviderContainer container;
  late InMemoryRefreshTokenStore store;

  setUp(() {
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
  });

  group(
    'session against the contract mock',
    () {
      test('login returns a session whose tokens are kept', () async {
        final auth = await container
            .read(authRepositoryProvider)
            .login(
              email: 'ana.gomez@example.mx',
              password: syntheticPassword(),
            );
        final session = container.read(sessionProvider);

        await session.start(auth.tokens);

        expect(store.token, auth.tokens.refreshToken);
        expect(session.accessToken, auth.tokens.accessToken);
      });

      test('refresh rotates through the contract endpoint', () async {
        final session = container.read(sessionProvider);
        await session.start(tokenPair(access: 'a', refresh: 'r'));

        final access = await session.refresh();

        expect(access, isNotNull);
        expect(store.written, hasLength(2), reason: 'rotation stored');
      });

      test('logout is accepted and wipes the store', () async {
        final session = container.read(sessionProvider);
        await session.start(tokenPair(access: 'a', refresh: 'r'));

        await container.read(signOutProvider).call();

        expect(store.token, isNull);
        expect(session.status, SessionStatus.signedOut);
      });
    },
    skip: _mockUrl.isEmpty ? 'set MOCK_API_URL (make mobile-test-mock)' : false,
  );
}
