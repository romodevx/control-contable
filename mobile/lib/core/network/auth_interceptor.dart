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

/// Requests that re-enter the current password (step-up re-auth). Their
/// 401 means a wrong password, not an expired bearer: a refresh and a
/// retry would submit the same wrong password twice (two strikes toward
/// the login failure limit), so their 401 is answered as is. The caller
/// makes sure the bearer is fresh just before (`DeleteAccount`).
const Set<(String, String)> _reauthRequests = {('DELETE', '/v1/me')};

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
/// - Auth endpoints (`/v1/auth/*`) and re-auth requests are never
///   refreshed for.
/// - Only requests to the configured API origin (scheme, host and port of
///   `apiBase`) get the bearer or a refresh: an absolute URL to anywhere
///   else never sees the token.
///
/// A plain [Interceptor], not a `QueuedInterceptor`: the single flight lives
/// in `Session`, so a retry through the same client never waits behind its
/// own error handler.
class AuthInterceptor extends Interceptor {
  /// Authenticates the requests of a client with a session.
  new({required this._session, required this._dio, required this._apiBase});

  final Session _session;
  final Dio _dio;
  final Uri _apiBase;

  @override
  Future<void> onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    final uri = options.uri;
    final path = uri.path;
    if (!_isApiOrigin(uri) || _publicAuthPaths.any(path.endsWith)) {
      return handler.next(options);
    }
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
        _isApiOrigin(request.uri) &&
        !_isAuthPath(request.uri.path) &&
        !_isReauth(request) &&
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

  bool _isApiOrigin(Uri uri) =>
      uri.scheme == _apiBase.scheme &&
      uri.host.toLowerCase() == _apiBase.host.toLowerCase() &&
      uri.port == _apiBase.port;

  static bool _isAuthPath(String path) => path.contains(_authPrefix);

  static bool _isReauth(RequestOptions request) => _reauthRequests.any(
    (reauth) =>
        reauth.$1 == request.method.toUpperCase() &&
        request.uri.path.endsWith(reauth.$2),
  );

  static Object? _bearerOf(RequestOptions options) =>
      options.headers[HttpHeaders.authorizationHeader];

  static void _setBearer(RequestOptions options, String token) =>
      options.headers[HttpHeaders.authorizationHeader] = 'Bearer $token';
}
