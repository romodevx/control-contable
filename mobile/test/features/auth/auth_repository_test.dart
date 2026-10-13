import 'dart:convert';

import 'package:control_contable/api/generated/api_client.dart';
import 'package:control_contable/api/generated/models/document_acceptance.dart';
import 'package:control_contable/api/generated/models/legal_document_kind.dart';
import 'package:control_contable/api/generated/models/register_request.dart';
import 'package:control_contable/api/generated/models/verify_email_request.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/auth/data/auth_repository.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/recording_adapter.dart';

const CannedReply _verificationSent = (
  status: 202,
  body: <String, Object?>{'status': 'verification_sent'},
);

void main() {
  late RecordingAdapter adapter;
  late ApiAuthRepository repository;

  setUp(() {
    adapter = RecordingAdapter();
    final dio = Dio(BaseOptions(baseUrl: 'http://localhost:3001'))
      ..httpClientAdapter = adapter;
    repository = ApiAuthRepository(ApiClient(dio).auth);
  });

  Map<String, Object?> lastBody() =>
      jsonDecode(jsonEncode(adapter.requests.last.data))
          as Map<String, Object?>;

  test('register posts the request and accepts the generic 202', () async {
    adapter.reply('/v1/auth/register', _verificationSent);

    await repository.register(
      const RegisterRequest(
        email: 'ana@example.mx',
        password: 'una frase larga',
        acceptedDocuments: [
          DocumentAcceptance(
            kind: LegalDocumentKind.terms,
            version: '2026-10-01',
          ),
        ],
      ),
    );

    expect(adapter.requests.single.method, 'POST');
    expect(lastBody(), {
      'email': 'ana@example.mx',
      'password': 'una frase larga',
      'accepted_documents': [
        {'kind': 'terms', 'version': '2026-10-01'},
      ],
    });
  });

  test('a rejected register surfaces the field details', () async {
    adapter.reply('/v1/auth/register', (
      status: 400,
      body: {
        'error': {
          'code': 'validation_failed',
          'message': 'One or more fields are invalid.',
          'request_id': 'req_1',
          'details': [
            {
              'field': 'password',
              'code': 'password_policy',
              'message': 'Use at least 12 characters.',
            },
          ],
        },
      },
    ));

    await expectLater(
      repository.register(
        const RegisterRequest(
          email: 'ana@example.mx',
          password: 'password1234',
          acceptedDocuments: [],
        ),
      ),
      throwsA(
        isA<ApiValidationFailure>().having(
          (f) => f.details.single.code,
          'code',
          'password_policy',
        ),
      ),
    );
  });

  test('verifyEmail returns the session', () async {
    adapter.reply('/v1/auth/verify-email', (
      status: 201,
      body: {
        'user': {
          'id': '7a1f0c62-1b1e-4c1d-9a55-0f1e2d3c4b5a',
          'email': 'ana@example.mx',
          'created_at': '2026-10-13T12:00:00Z',
        },
        'tokens': {
          'access_token': 'a',
          'access_expires_in': 900,
          'refresh_token': 'r',
          'refresh_expires_in': 2592000,
          'token_type': 'Bearer',
        },
      },
    ));

    final session = await repository.verifyEmail(
      const VerifyEmailRequest(email: 'ana@example.mx', code: '482913'),
    );

    expect(session.user.email, 'ana@example.mx');
    expect(lastBody(), {'email': 'ana@example.mx', 'code': '482913'});
  });

  test('a wrong code is unauthorized', () async {
    adapter.reply('/v1/auth/verify-email', (status: 401, body: null));

    await expectLater(
      repository.verifyEmail(
        const VerifyEmailRequest(email: 'ana@example.mx', code: '000000'),
      ),
      throwsA(isA<ApiUnauthorized>()),
    );
  });

  test('resendVerification posts the email', () async {
    adapter.reply('/v1/auth/verify-email/resend', _verificationSent);

    await repository.resendVerification('ana@example.mx');

    expect(lastBody(), {'email': 'ana@example.mx'});
  });

  test('listLegalDocuments returns the items', () async {
    adapter.reply('/v1/legal/documents', (
      status: 200,
      body: {
        'items': [
          {
            'kind': 'terms',
            'version': '2026-10-01',
            'url': 'https://example.mx/legal/terminos',
            'published_at': '2026-10-01T06:00:00Z',
          },
        ],
      },
    ));

    final documents = await repository.listLegalDocuments();

    expect(documents.single.kind, LegalDocumentKind.terms);
  });
}
