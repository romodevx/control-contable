import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/core/network/https_only_interceptor.dart';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

const _connectTimeout = Duration(seconds: 10);
const _receiveTimeout = Duration(seconds: 20);

/// The single HTTP client. The generated API client (`apiClientProvider`) and
/// the auth interceptors (M1) are built on top of this instance.
///
/// Redirects are never followed: the API does not redirect, and a followed
/// 3xx could downgrade HTTPS to HTTP inside the platform `HttpClient`, out of
/// reach of interceptors. A 3xx surfaces as a `DioException` instead. When
/// the base URL is HTTPS, `HttpsOnlyInterceptor` also refuses any `http://`
/// request.
final dioProvider = Provider<Dio>((ref) {
  final config = ref.watch(appConfigProvider);
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
  ref.onDispose(dio.close);
  return dio;
}, name: 'dioProvider');
