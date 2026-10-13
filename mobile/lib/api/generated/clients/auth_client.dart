// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

import '../models/auth_session.dart';
import '../models/legal_document_list.dart';
import '../models/login_request.dart';
import '../models/refresh_request.dart';
import '../models/register_accepted.dart';
import '../models/register_request.dart';
import '../models/token_pair.dart';
import '../models/user.dart';
import '../models/user_update.dart';
import '../models/verify_email_request.dart';

part 'auth_client.g.dart';

@RestApi()
abstract class AuthClient {
  factory AuthClient(Dio dio, {String? baseUrl}) = _AuthClient;

  /// Start registration (sends a verification code).
  ///
  /// Always answers `202` with the same body, whether or not the email exists,.
  /// so accounts cannot be enumerated (US-A1). Tokens are issued by.
  /// `verifyEmail`. Requires acceptance of the current legal documents.
  /// (ADR-0001 OWNER decision 1, option A.).
  @POST('/v1/auth/register')
  Future<RegisterAccepted> register({@Body() required RegisterRequest body});

  /// Confirm the emailed code; creates the account, starts the trial, issues tokens
  @POST('/v1/auth/verify-email')
  Future<AuthSession> verifyEmail({@Body() required VerifyEmailRequest body});

  /// Log in with email and password
  @POST('/v1/auth/login')
  Future<AuthSession> login({@Body() required LoginRequest body});

  /// Rotate the refresh token.
  ///
  /// Reusing an already-rotated token revokes its whole family and returns 401.
  @POST('/v1/auth/refresh')
  Future<TokenPair> refreshSession({@Body() required RefreshRequest body});

  /// Revoke the refresh token family
  @POST('/v1/auth/logout')
  Future<void> logout({@Body() required RefreshRequest body});

  /// Current user profile
  @GET('/v1/me')
  Future<User> getMe();

  /// Update profile
  @PATCH('/v1/me')
  Future<User> updateMe({@Body() required UserUpdate body});

  /// Delete account and all its data (irreversible).
  ///
  /// Deletes the user, taxpayers and records (cascade) and revokes all sessions.
  /// Does not cancel the store subscription (the app explains how). The keyed.
  /// email hash stays in the trial registry (no second trial).
  @DELETE('/v1/me')
  Future<void> deleteMe();

  /// Current privacy notices and terms
  @GET('/v1/legal/documents')
  Future<LegalDocumentList> listLegalDocuments();
}
