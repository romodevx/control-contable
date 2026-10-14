import 'dart:io';

import 'package:control_contable/core/session/session.dart';
import 'package:dio/dio.dart';

/// The auth endpoints' common prefix. A 401 from any of them is the answer
/// itself (bad credentials, bad code, bad or reused refresh token), never a
/// reason to refresh.
const String _authPrefix = '/v1/auth/';

/// Auth endpoints that take no bearer (every one except `logout`).
const Set<String> _publicAuthPaths = {
  '/v1/auth/login',
  '/v1/auth/refresh',
  '/v1/auth/register',
  '/v1/auth/verify-email',
  '/v1/auth/verify-email/resend',
};

/// Marks a request already retried after a refresh: it is never refreshed
/// for again, so a second 401 cannot loop.
const String _retriedKey = 'auth.retried';

/// Adds the session's bearer to each request and answers a 401 with one
/// refresh and one retry.
///
/// - The bearer is the in-memory access token. After a cold start (a
///   refresh token kept, no access token yet) the first request refreshes
///   before it leaves instead of collecting a 401.
/// - On a 401 the request is retried once: with the current token when a
///   refresh already replaced the one it carried, else after
///   `Session.refresh`, which is single-flight (concurrent 401s share one
///   rotation) and stores the rotated refresh token.
/// - A refused refresh ends the session (the router then shows login) and
///   the original 401 surfaces; an unreachable one keeps the session.
/// - Auth endpoints (`/v1/auth/*`) are never refreshed for.
///
/// A plain [Interceptor], not a `QueuedInterceptor`: the single flight lives
/// in `Session`, so a retry through the same client never waits behind its
/// own error handler.
class AuthInterceptor extends Interceptor {
  /// Authenticates the requests of a client with a session.
  new({required this._session, required this._dio});

  final Session _session;
  final Dio _dio;

  @override
  Future<void> onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    final path = options.uri.path;
    if (_publicAuthPaths.any(path.endsWith)) return handler.next(options);
    final mustRefreshFirst =
        _session.isSignedIn &&
        _session.accessToken == null &&
        !_isAuthPath(path);
    final token = mustRefreshFirst
        ? await _session.refresh()
        : _session.accessToken;
    if (token != null) _setBearer(options, token);
    handler.next(options);
  }

  @override
  Future<void> onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    final request = err.requestOptions;
    final isRefreshable =
        err.response?.statusCode == HttpStatus.unauthorized &&
        !_isAuthPath(request.uri.path) &&
        request.extra[_retriedKey] != true &&
        _session.isSignedIn;
    if (!isRefreshable) return handler.next(err);

    final current = _session.accessToken;
    final replaced = current != null && _bearerOf(request) != 'Bearer $current';
    final token = replaced ? current : await _session.refresh();
    if (token == null) return handler.next(err);

    // A copy: the failed request stays as it was sent.
    final retry = request.copyWith(
      headers: {...request.headers},
      extra: {...request.extra, _retriedKey: true},
    );
    _setBearer(retry, token);
    try {
      handler.resolve(await _dio.fetch<dynamic>(retry));
    } on DioException catch (retryError) {
      handler.next(retryError);
    }
  }

  static bool _isAuthPath(String path) => path.contains(_authPrefix);

  static Object? _bearerOf(RequestOptions options) =>
      options.headers[HttpHeaders.authorizationHeader];

  static void _setBearer(RequestOptions options, String token) =>
      options.headers[HttpHeaders.authorizationHeader] = 'Bearer $token';
}
