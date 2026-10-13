// Integration test against the Prism mock (`make mock`; see
// docs/runbooks/mock-server.md): the sign-up calls the register and verify
// screens make, through the generated client, validated by the contract.
// Runs only when MOCK_API_URL is set: `make mobile-test-mock` from the repo
// root.
@Tags(['mock'])
library;

import 'dart:io';

import 'package:control_contable/api/generated/models/register_request.dart';
import 'package:control_contable/api/generated/models/verify_email_request.dart';
import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/features/auth/application/legal_documents.dart';
import 'package:control_contable/features/auth/data/auth_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../helpers/synthetic_password.dart';

final String _mockUrl = Platform.environment['MOCK_API_URL'] ?? '';

void main() {
  late ProviderContainer container;
  late AuthRepository repository;

  setUp(() {
    container = ProviderContainer.test(
      overrides: [
        appConfigProvider.overrideWithValue(
          AppConfig.resolve(
            flavor: AppFlavor.dev,
            apiBaseUrlOverride: _mockUrl,
          ),
        ),
      ],
    );
    repository = container.read(authRepositoryProvider);
  });

  group(
    'sign-up against the contract mock',
    () {
      test('the documents yield an acceptable set', () async {
        final documents = await repository.listLegalDocuments();

        final set = LegalDocumentSet.pick(documents);
        expect(set, isNotNull);
        expect(Uri.parse(set!.terms.url).isScheme('https'), isTrue);
      });

      test(
        'register without a name is accepted (no null on the wire)',
        () async {
          final set = LegalDocumentSet.pick(
            await repository.listLegalDocuments(),
          )!;

          // Prism rejects a body that breaks the schema (400 with
          // sl-violations); a 202 means the request matched the contract.
          await repository.register(
            RegisterRequest(
              email: 'ana.gomez@example.mx',
              password: syntheticPassword(24),
              acceptedDocuments: set.acceptances,
            ),
          );
        },
      );

      test('verify returns a session and resend the generic 202', () async {
        final session = await repository.verifyEmail(
          const VerifyEmailRequest(
            email: 'ana.gomez@example.mx',
            code: '482913',
          ),
        );

        expect(session.tokens.refreshToken, isNotEmpty);
        await repository.resendVerification('ana.gomez@example.mx');
      });
    },
    skip: _mockUrl.isEmpty ? 'set MOCK_API_URL (make mobile-test-mock)' : false,
  );
}
