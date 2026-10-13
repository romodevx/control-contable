import 'package:dio/dio.dart';

const _https = 'https';

/// Raised (inside a [DioException]) when an HTTPS client is asked to send a
/// request over plain HTTP.
class InsecureRequestError extends Error {
  /// Creates the error for the refused [uri].
  new(this.uri);

  /// The request that was refused.
  final Uri uri;

  @override
  String toString() => 'InsecureRequestError: refused non-HTTPS request $uri';
}

/// Rejects every request whose final URI is not HTTPS, before it leaves the
/// device. Installed on staging and prod, whose base URL is HTTPS.
///
/// Dart sockets ignore Android Network Security Config and iOS ATS, so the
/// app has to enforce HTTPS itself. Redirects are not followed at all
/// (`followRedirects: false` in `dioProvider`), so a 3xx cannot move a
/// request to `http://` either.
class HttpsOnlyInterceptor extends Interceptor {
  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    if (options.uri.scheme == _https) {
      handler.next(options);
      return;
    }
    handler.reject(
      DioException(
        requestOptions: options,
        error: InsecureRequestError(options.uri),
        message: 'Only HTTPS requests are allowed',
      ),
    );
  }
}
