import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/core/network/api_dio.dart';
import 'package:control_contable/core/network/auth_interceptor.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The single HTTP client the app's calls share (`apiClientProvider` is
/// built on it): the hardened transport of [buildApiDio] plus the session's
/// [AuthInterceptor] (bearer on each request, one refresh on a 401).
final dioProvider = Provider<Dio>((ref) {
  final dio = buildApiDio(ref.watch(appConfigProvider));
  dio.interceptors.add(
    AuthInterceptor(session: ref.watch(sessionProvider), dio: dio),
  );
  ref.onDispose(dio.close);
  return dio;
}, name: 'dioProvider');
