import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/core/network/https_only_interceptor.dart';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

const _connectTimeout = Duration(seconds: 10);
const _receiveTimeout = Duration(seconds: 20);

/// A [Dio] for the API at [config]'s base URL, with the transport
/// hardening every client shares.
///
/// Redirects are never followed: the API does not redirect, and a followed
/// 3xx could downgrade HTTPS to HTTP inside the platform `HttpClient`, out of
/// reach of interceptors. A 3xx surfaces as a `DioException` instead. When
/// the base URL is HTTPS, `HttpsOnlyInterceptor` also refuses any `http://`
/// request.
Dio buildApiDio(AppConfig config) {
  final dio = Dio(
    BaseOptions(
      baseUrl: config.apiBaseUrl.toString(),
      connectTimeout: _connectTimeout,
      receiveTimeout: _receiveTimeout,
      contentType: Headers.jsonContentType,
      followRedirects: false,
      maxRedirects: 0,
    ),
  );
  if (config.apiBaseUrl.isScheme('https')) {
    dio.interceptors.add(HttpsOnlyInterceptor());
  }
  return dio;
}

/// A client with no session behaviour: no bearer, no refresh on 401. Only
/// the token rotation uses it, so a refused refresh can never recurse into
/// another refresh. Everything else goes through `dioProvider`.
final bareDioProvider = Provider<Dio>((ref) {
  final dio = buildApiDio(ref.watch(appConfigProvider));
  ref.onDispose(dio.close);
  return dio;
}, name: 'bareDioProvider');
