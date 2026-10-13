import 'dart:math' as math;

import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/l10n/app_localizations.dart';

const int _secondsPerMinute = 60;

/// New-password length limits (`VerifyEmailRequest.password`).
const int _passwordMinLength = 12;
const int _passwordMaxLength = 128;

/// The client check of a new password: its length, counted in characters.
/// The common/breached check runs on the server (D5).
FieldError? checkNewPassword(String value, AppLocalizations l10n) {
  final length = value.runes.length;
  if (length < _passwordMinLength) {
    return CustomFieldError(l10n.authPasswordTooShort);
  }
  if (length > _passwordMaxLength) {
    return CustomFieldError(l10n.authPasswordTooLong);
  }
  return null;
}

/// Server `FieldError.code` values with their own copy.
const String _passwordPolicyCode = 'password_policy';
const String _tooShortCode = 'minLength';
const String _tooLongCode = 'maxLength';

/// The es-MX copy for a sign-up field [detail] (`register` or
/// `verifyEmail`), keyed by the root of its dotted path and its code. The
/// server's English `message` is never shown.
FieldError describeSignUpDetail(api.FieldError detail, AppLocalizations l10n) {
  final root = detail.field.split('.').first;
  final message = switch ((root, detail.code)) {
    ('password', _passwordPolicyCode) => l10n.authPasswordPolicy,
    ('password', _tooShortCode) => l10n.authPasswordTooShort,
    ('password', _tooLongCode) => l10n.authPasswordTooLong,
    ('email', _) => l10n.authEmailInvalid,
    ('name', _) => l10n.authNameTooLong,
    ('code', _) => l10n.authCodeFormat,
    ('accepted_documents', _) => l10n.authDocumentsStale,
    _ => l10n.authFieldRejected,
  };
  return CustomFieldError(message);
}

/// The form-level es-MX message for [failure]. Field verdicts
/// (`ApiValidationFailure` details) are shown on their fields instead.
String describeAuthFailure(ApiFailure failure, AppLocalizations l10n) =>
    switch (failure) {
      ApiRateLimited(:final retryAfter?) => l10n.authErrorRateLimited(
        _minutesRoundedUp(retryAfter),
      ),
      ApiRateLimited() => l10n.authErrorRateLimitedNoWait,
      ApiUnreachable() => l10n.authErrorUnreachable,
      ApiNotFound() => l10n.authErrorUnavailable,
      ApiValidationFailure() => l10n.authFormRejected,
      ApiUnauthorized() || ApiUnexpected() => l10n.authErrorUnexpected,
    };

/// A countdown as `m:ss` (`1:00`, `0:45`, `12:05`).
String formatCountdown(Duration remaining) {
  final seconds = remaining.inSeconds % _secondsPerMinute;
  return '${remaining.inMinutes}:${seconds.toString().padLeft(2, '0')}';
}

int _minutesRoundedUp(Duration wait) =>
    math.max(1, (wait.inSeconds / _secondsPerMinute).ceil());
