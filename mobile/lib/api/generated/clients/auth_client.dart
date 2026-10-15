// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

import '../models/auth_session.dart';
import '../models/delete_account_request.dart';
import '../models/legal_document_list.dart';
import '../models/login_request.dart';
import '../models/logout_request.dart';
import '../models/refresh_request.dart';
import '../models/register_request.dart';
import '../models/resend_verification_request.dart';
import '../models/token_pair.dart';
import '../models/user.dart';
import '../models/user_update.dart';
import '../models/verification_sent.dart';
import '../models/verify_email_request.dart';

part 'auth_client.g.dart';

@RestApi()
abstract class AuthClient {
  factory AuthClient(Dio dio, {String? baseUrl}) = _AuthClient;

  /// Start registration (emails a 6-digit verification code).
  ///
  /// Always answers `202 {status: verification_sent}` with the same body and comparable latency, whether the address is new, has a pending sign-up or already belongs to a verified account, so accounts cannot be enumerated (US-A1a). For a verified account no code is created and the owner gets a "someone tried to sign up with your email" notice, at most once per hour (D3). Register takes no password and no name: both are set at `verifyEmail` by whoever proves they own the inbox, so pre-registering someone else's address gives the attacker nothing (D8). For an email with a live (unexpired) pending sign-up the stored accepted documents are never changed: the call behaves like `resendVerification` for that email (same generic 202, same cooldown and rate limits) (D7). Only once the pending sign-up has expired does a new register replace it. Nothing is written to users, consents or entitlements until `verifyEmail` (ADR-0001 OWNER decision 1, option A).
  /// `accepted_documents` must carry the current version (from `listLegalDocuments`) of `terms` and of one privacy notice (`privacy_notice_integral` or `privacy_notice_simplified`). A stale version or a missing kind answers `400 validation_failed` on `accepted_documents`.
  @POST('/v1/auth/register')
  Future<VerificationSent> register({@Body() required RegisterRequest body});

  /// Confirm the emailed code and set the password; creates the account, starts the trial, issues tokens.
  ///
  /// The password (and the optional name) is chosen here, by whoever holds the emailed code, never at `register`: an attacker who pre-registers someone else's address cannot choose the password of the account the owner later verifies (D8).
  /// In one transaction: creates the user with this password, stores the accepted consents (kind, version, accepted_at), consumes the pending sign-up and starts the trial on the server clock only if the email's keyed hash is not already in the trial registry (an account deleted and created again gets no second trial; same `201`) (US-A1b).
  /// The body, the password policy and the name are checked before the code: a short, long, common or breached password answers `400 validation_failed` on `password` (the value is never echoed) and does not count as a try. A wrong, expired, consumed or unknown code, or an unknown email, answers `401` with one generic message. Each wrong try counts against the code; after 5 the code is dead and every try answers `401` until `resendVerification`.
  @POST('/v1/auth/verify-email')
  Future<AuthSession> verifyEmail({@Body() required VerifyEmailRequest body});

  /// Email a new verification code for a pending sign-up.
  ///
  /// For a pending sign-up, a new code replaces the old one (attempts reset). Always answers the same `202 {status: verification_sent}`: also when there is no pending sign-up, the email is already verified, or the last resend was less than 60 s ago (the cooldown is silent so it is not an oracle; the app shows a 60 s countdown) (US-A1c, D2). A `register` for an email with a live pending sign-up takes this same path and shares its cooldown and limits; neither call ever changes the pending sign-up's accepted documents (D7). A pending sign-up holds no password: it is set at `verifyEmail` (D8).
  @POST('/v1/auth/verify-email/resend')
  Future<VerificationSent> resendVerification({
    @Body() required ResendVerificationRequest body,
  });

  /// Log in with email and password (starts a new token family).
  ///
  /// A wrong password, an unknown email and a pending (unverified) sign-up all answer the same `401` body with comparable latency (US-A2a). Too many failures for the account or the client answer `429 rate_limited` with `Retry-After`.
  @POST('/v1/auth/login')
  Future<AuthSession> login({@Body() required LoginRequest body});

  /// Rotate the refresh token.
  ///
  /// A live refresh token is rotated: the response carries a new pair and the old token is marked rotated (US-A2b). Reusing an already-rotated token revokes its whole family and answers `401`; of two concurrent refreshes with the same token exactly one wins and the other is treated as reuse. An expired (30-day sliding or 90-day absolute), revoked, malformed or foreign token answers `401`.
  @POST('/v1/auth/refresh')
  Future<TokenPair> refreshSession({@Body() required RefreshRequest body});

  /// Revoke the refresh token family.
  ///
  /// Revokes the family of the given refresh token (US-A2c). Idempotent: a refresh token that is already revoked or not the caller's also answers `204`, so another user's token state is never revealed. A missing or invalid bearer answers `401`.
  ///
  /// **Push device (0.14.0).** The optional `device_id` is this installation's id (the `registerDevice` `id`). When given, the caller's device row with that id is removed in the same transaction as the revocation, so a phone that changes hands stops receiving the previous user's reminders even when the app's earlier `unregisterDevice` call was lost (M9 security gate F1). Another user's or an unknown id removes nothing and still answers `204` (no oracle). It is not gated by `module_reminders`: removing one's own device is always allowed, and with the module OFF there is nothing else to undo. A malformed id is `400 validation_failed` on `device_id`.
  @POST('/v1/auth/logout')
  Future<void> logout({@Body() required LogoutRequest body});

  /// Current user profile
  @GET('/v1/me')
  Future<User> getMe();

  /// Update profile (name only; email change is out of scope)
  @PATCH('/v1/me')
  Future<User> updateMe({@Body() required UserUpdate body});

  /// Delete the account and all its data (irreversible; requires the current password).
  ///
  /// Re-authenticates with the current password (D4), then in one transaction deletes the user and everything it owns (taxpayers, records, consents, refresh tokens, usage, entitlements) and answers `204` (US-A4). Consents are deleted too (D6); only the keyed email hash survives in the trial registry (no second trial). The store subscription is not cancelled (the app explains how). Any later call with the old access token answers `401`.
  /// A wrong password answers `403 invalid_password` and deletes nothing: the bearer was valid, so it is not a `401` (which always means the bearer, carries the challenge and is the client's cue to refresh). Each account gets 5 password attempts, then one more every 3 minutes (`429 rate_limited` with `Retry-After`); a body without `password` answers `400 validation_failed` and spends no attempt.
  @DELETE('/v1/me')
  Future<void> deleteMe({@Body() required DeleteAccountRequest body});

  /// Current privacy notices and terms.
  ///
  /// The current version of each legal document kind (US-A5). The register screen links to them and sends back the versions shown in `accepted_documents`. Versions are data, not code.
  @GET('/v1/legal/documents')
  Future<LegalDocumentList> listLegalDocuments();
}
