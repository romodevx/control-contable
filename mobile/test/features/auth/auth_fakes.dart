import 'dart:async';

import 'package:control_contable/api/generated/models/access_level.dart';
import 'package:control_contable/api/generated/models/access_summary.dart';
import 'package:control_contable/api/generated/models/auth_session.dart';
import 'package:control_contable/api/generated/models/legal_document.dart';
import 'package:control_contable/api/generated/models/legal_document_kind.dart';
import 'package:control_contable/api/generated/models/me.dart';
import 'package:control_contable/api/generated/models/register_request.dart';
import 'package:control_contable/api/generated/models/token_pair.dart';
import 'package:control_contable/api/generated/models/user.dart';
import 'package:control_contable/api/generated/models/verify_email_request.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/auth/data/auth_repository.dart';
import 'package:control_contable/features/auth/data/profile_repository.dart';

/// The contract's current documents (the `LegalDocumentList` example).
final List<LegalDocument> currentDocuments = [
  _document(LegalDocumentKind.privacyNoticeIntegral, 'aviso-integral'),
  _document(LegalDocumentKind.privacyNoticeSimplified, 'aviso-simplificado'),
  _document(LegalDocumentKind.terms, 'terminos'),
];

LegalDocument _document(LegalDocumentKind kind, String slug) => LegalDocument(
  kind: kind,
  version: '2026-10-01',
  url: 'https://example.mx/legal/$slug',
  publishedAt: DateTime.utc(2026, 10),
);

/// A session as `verifyEmail` answers it.
final AuthSession sampleSession = AuthSession(
  user: User(
    id: '7a1f0c62-1b1e-4c1d-9a55-0f1e2d3c4b5a',
    email: 'ana@example.mx',
    createdAt: DateTime.utc(2026, 10, 13),
  ),
  tokens: const TokenPair(
    accessToken: 'access',
    accessExpiresIn: 900,
    refreshToken: 'refresh',
    refreshExpiresIn: 2592000,
    tokenType: 'Bearer',
  ),
);

/// An [AuthRepository] scripted per test: each call is recorded, and a
/// `*Error` makes that call throw it. A `*Gate` holds the call open until
/// the test completes it (to observe the busy state).
class FakeAuthRepository implements AuthRepository {
  List<LegalDocument> documents = currentDocuments;
  Exception? documentsError;
  int documentCalls = 0;

  final List<RegisterRequest> registered = [];
  Exception? registerError;
  Completer<void>? registerGate;

  final List<VerifyEmailRequest> verified = [];
  Exception? verifyError;

  final List<String> resent = [];
  Exception? resendError;

  final List<({String email, String password})> logins = [];
  Exception? loginError;
  Completer<void>? loginGate;

  final List<String> loggedOut = [];

  /// The push device id each logout carried (null when none).
  final List<String?> loggedOutDevices = [];
  Exception? logoutError;

  /// Called on each logout, before it answers (to observe ordering).
  void Function()? onLogout;

  @override
  Future<List<LegalDocument>> listLegalDocuments() async {
    documentCalls++;
    if (documentsError case final error?) throw error;
    return documents;
  }

  @override
  Future<void> register(RegisterRequest request) async {
    registered.add(request);
    await registerGate?.future;
    if (registerError case final error?) throw error;
  }

  @override
  Future<AuthSession> verifyEmail(VerifyEmailRequest request) async {
    verified.add(request);
    if (verifyError case final error?) throw error;
    return sampleSession;
  }

  @override
  Future<void> resendVerification(String email) async {
    resent.add(email);
    if (resendError case final error?) throw error;
  }

  @override
  Future<AuthSession> login({
    required String email,
    required String password,
  }) async {
    logins.add((email: email, password: password));
    await loginGate?.future;
    if (loginError case final error?) throw error;
    return sampleSession;
  }

  @override
  Future<void> logout(String refreshToken, {String? deviceId}) async {
    onLogout?.call();
    loggedOut.add(refreshToken);
    loggedOutDevices.add(deviceId);
    if (logoutError case final error?) throw error;
  }
}

/// The profile `GET /v1/me` answers by default.
final User sampleUser = User(
  id: '7a1f0c62-1b1e-4c1d-9a55-0f1e2d3c4b5a',
  email: 'ana@example.mx',
  name: 'Ana Gómez',
  createdAt: DateTime.utc(2026, 10, 13),
);

/// The access `GET /v1/me` carries by default: a fresh trial.
final AccessSummary sampleAccess = AccessSummary(
  access: AccessLevel.trial,
  trialEndsAt: DateTime.utc(2026, 10, 20),
  trialDaysLeft: 7,
  incomesRemaining: 5,
  expensesRemaining: 5,
);

/// A [ProfileRepository] scripted per test, like [FakeAuthRepository]. The
/// calls are recorded in [calls] in order (`getMe`, `updateName`,
/// `deleteAccount`).
class FakeProfileRepository implements ProfileRepository {
  User user = sampleUser;
  AccessSummary access = sampleAccess;
  final List<String> calls = [];

  Exception? getMeError;
  Completer<void>? getMeGate;

  final List<String> names = [];
  Exception? updateError;
  Completer<void>? updateGate;

  final List<String> deletePasswords = [];
  Exception? deleteError;
  Completer<void>? deleteGate;

  @override
  Future<Me> getMe() async {
    calls.add('getMe');
    await getMeGate?.future;
    if (getMeError case final error?) throw error;
    return Me(
      id: user.id,
      email: user.email,
      name: user.name,
      createdAt: user.createdAt,
      access: access,
    );
  }

  @override
  Future<User> updateName(String name) async {
    calls.add('updateName');
    names.add(name);
    await updateGate?.future;
    if (updateError case final error?) throw error;
    return user = user.copyWith(name: name);
  }

  @override
  Future<void> deleteAccount(String password) async {
    calls.add('deleteAccount');
    deletePasswords.add(password);
    await deleteGate?.future;
    if (deleteError case final error?) throw error;
  }
}

/// The refresh token a signed-in harness starts with.
const String storedRefreshToken = 'refresh-stored';

/// A scripted `refreshSession`: counts calls and answers a fresh pair, or
/// throws [error].
class FakeTokenRotation {
  final List<String> received = [];
  ApiFailure? error;

  Future<TokenPair> call(String refreshToken) async {
    received.add(refreshToken);
    if (error case final error?) throw error;
    return const TokenPair(
      accessToken: 'access-rotated',
      accessExpiresIn: 900,
      refreshToken: 'refresh-rotated',
      refreshExpiresIn: 2592000,
      tokenType: 'Bearer',
    );
  }
}

/// A clock the test moves by hand.
class FakeClock {
  DateTime now = DateTime.utc(2026, 10, 13, 12);

  DateTime call() => now;

  void advance(Duration by) => now = now.add(by);
}
