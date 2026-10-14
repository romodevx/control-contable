import 'dart:io';

import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/api/generated/models/error_response.dart';
import 'package:control_contable/api/generated/models/field_error.dart';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

/// Why an API call failed, in the terms a screen acts on. Built from the
/// single `ErrorResponse` envelope of the contract; the server's English
/// `message` is never shown (screens own their es-MX copy).
@immutable
sealed class ApiFailure implements Exception {
  const new();
}

/// `400 validation_failed`: [details] name the rejected fields by their
/// dotted path (`accepted_documents.1.kind`).
final class ApiValidationFailure extends ApiFailure {
  /// The failure with the envelope's field [details].
  const new(this.details);

  /// One entry per rejected field; empty when the body had none.
  final List<FieldError> details;
}

/// `401`: missing or bad credentials, code or token (one generic answer).
final class ApiUnauthorized extends ApiFailure {
  /// The failure.
  const new();
}

/// `invalid_password` (a `403` on re-authentication, e.g. `DELETE /v1/me`):
/// the bearer was fine, the current password was not. Read by its code,
/// whatever the status, and never a cue to refresh.
final class ApiInvalidPassword extends ApiFailure {
  /// The failure.
  const new();
}

/// `404`: not found, or the module is off (indistinguishable by design).
final class ApiNotFound extends ApiFailure {
  /// The failure.
  const new();
}

/// `429`: too many tries; wait [retryAfter] (when the server said).
final class ApiRateLimited extends ApiFailure {
  /// The failure with the server's wait, if any.
  const new(this.retryAfter);

  /// How long to wait before trying again.
  final Duration? retryAfter;
}

/// The server could not be reached (offline, timeout).
final class ApiUnreachable extends ApiFailure {
  /// The failure.
  const new();
}

/// Anything else: a 5xx, an unknown status, an unreadable body.
final class ApiUnexpected extends ApiFailure {
  /// The failure.
  const new();
}

const Set<DioExceptionType> _unreachableTypes = {
  DioExceptionType.connectionError,
  DioExceptionType.connectionTimeout,
  DioExceptionType.receiveTimeout,
  DioExceptionType.sendTimeout,
};

/// Runs [call] (a generated client method) and rethrows its failure as an
/// [ApiFailure]; [now] dates an HTTP-date `Retry-After`.
///
/// A body that does not decode into the contract's shape is an
/// [ApiUnexpected], whatever the decoder threw: a `FormatException` or a
/// `CheckedFromJsonException` (both `Exception`s), or the `TypeError` of a
/// generated cast on a null, missing or wrongly typed field (an `Error`,
/// M1-16 FU1). The decoder's error is dropped on purpose: its text can
/// quote the payload (a token). Any other `Error` is a bug and propagates.
Future<T> guardApi<T>(
  Future<T> Function() call, {
  DateTime Function() now = DateTime.now,
}) async {
  try {
    return await call();
  } on DioException catch (exception) {
    throw apiFailureFrom(exception, now: now);
  } on Exception {
    throw const ApiUnexpected();
    // The generated decoders signal a contract mismatch only by a failed
    // cast: at this boundary that is bad input, not a bug (see above).
    // ignore: avoid_catching_errors
  } on TypeError {
    throw const ApiUnexpected();
  }
}

/// The [ApiFailure] behind [exception].
ApiFailure apiFailureFrom(
  DioException exception, {
  DateTime Function() now = DateTime.now,
}) {
  if (_unreachableTypes.contains(exception.type)) return const ApiUnreachable();
  final response = exception.response;
  if (_codeOf(response?.data) == ErrorCode.invalidPassword) {
    return const ApiInvalidPassword();
  }
  return switch (response?.statusCode) {
    HttpStatus.badRequest => ApiValidationFailure(_details(response!.data)),
    HttpStatus.unauthorized => const ApiUnauthorized(),
    HttpStatus.notFound => const ApiNotFound(),
    HttpStatus.tooManyRequests => ApiRateLimited(
      parseRetryAfter(
        response!.headers.value(HttpHeaders.retryAfterHeader),
        now: now(),
      ),
    ),
    _ => const ApiUnexpected(),
  };
}

/// The wait a `Retry-After` header asks for: delta-seconds or an HTTP-date
/// relative to [now]. `null` when absent, unreadable or already past.
Duration? parseRetryAfter(String? header, {required DateTime now}) {
  if (header == null) return null;
  final seconds = int.tryParse(header.trim());
  if (seconds != null) return seconds > 0 ? Duration(seconds: seconds) : null;
  try {
    final wait = HttpDate.parse(header).difference(now);
    return wait > Duration.zero ? wait : null;
  } on HttpException {
    return null;
  }
}

ErrorCode? _codeOf(Object? body) {
  if (body is! Map<String, Object?>) return null;
  try {
    return ErrorResponse.fromJson(body).error.code;
  } on Object {
    // Not the contract's envelope: the status alone decides.
    return null;
  }
}

List<FieldError> _details(Object? body) {
  if (body is! Map<String, Object?>) return const [];
  try {
    return ErrorResponse.fromJson(body).error.details ?? const [];
  } on Object {
    // An envelope that does not match the contract carries no usable detail.
    return const [];
  }
}
