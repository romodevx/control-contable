import 'package:control_contable/api/generated/clients/auth_client.dart';
import 'package:control_contable/api/generated/models/auth_session.dart';
import 'package:control_contable/api/generated/models/legal_document.dart';
import 'package:control_contable/api/generated/models/register_request.dart';
import 'package:control_contable/api/generated/models/resend_verification_request.dart';
import 'package:control_contable/api/generated/models/verify_email_request.dart';
import 'package:control_contable/core/network/api_client_provider.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The sign-up operations of `module_auth`. Every method throws an
/// `ApiFailure` when the server refuses or cannot be reached.
abstract interface class AuthRepository {
  /// The current version of each legal document (`GET /v1/legal/documents`).
  Future<List<LegalDocument>> listLegalDocuments();

  /// Starts a sign-up; the server always answers the same generic 202.
  Future<void> register(RegisterRequest request);

  /// Confirms the emailed code; creates the account and its session.
  Future<AuthSession> verifyEmail(VerifyEmailRequest request);

  /// Asks for a new code; the server always answers the same generic 202.
  Future<void> resendVerification(String email);
}

/// [AuthRepository] over the generated auth client.
class ApiAuthRepository implements AuthRepository {
  /// The repository on [_client].
  new(this._client);

  final AuthClient _client;

  @override
  Future<List<LegalDocument>> listLegalDocuments() async =>
      (await guardApi(_client.listLegalDocuments)).items;

  @override
  Future<void> register(RegisterRequest request) =>
      guardApi(() => _client.register(body: request));

  @override
  Future<AuthSession> verifyEmail(VerifyEmailRequest request) =>
      guardApi(() => _client.verifyEmail(body: request));

  @override
  Future<void> resendVerification(String email) => guardApi(
    () => _client.resendVerification(
      body: ResendVerificationRequest(email: email),
    ),
  );
}

/// The auth operations. Tests override it with a fake.
final authRepositoryProvider = Provider<AuthRepository>(
  (ref) => ApiAuthRepository(ref.watch(apiClientProvider).auth),
  name: 'authRepositoryProvider',
);
