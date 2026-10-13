import 'dart:io';

import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/core/network/dio_provider.dart';
import 'package:control_contable/core/network/https_only_interceptor.dart';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/recording_adapter.dart';

Dio _dioFor(AppConfig config) {
  final container = ProviderContainer.test(
    overrides: [appConfigProvider.overrideWithValue(config)],
  );
  return container.read(dioProvider);
}

void main() {
  test('dio targets the configured API base URL', () {
    final dio = _dioFor(
      AppConfig.resolve(flavor: AppFlavor.dev, useMock: true),
    );

    expect(dio.options.baseUrl, 'http://localhost:3001');
    expect(dio.options.connectTimeout, isNotNull);
    expect(dio.options.receiveTimeout, isNotNull);
  });

  test('reading the config without an override fails loudly', () {
    final container = ProviderContainer.test();

    expect(() => container.read(appConfigProvider), throwsA(anything));
  });

  group('redirects (PR #15 follow-up: no HTTPS to HTTP downgrade)', () {
    late HttpServer server;
    late List<String> served;

    setUp(() async {
      served = [];
      server = await HttpServer.bind(InternetAddress.loopbackIPv4, 0)
        ..listen((request) async {
          served.add(request.uri.path);
          if (request.uri.path == '/v1/flags') {
            request.response
              ..statusCode = HttpStatus.found
              ..headers.set(HttpHeaders.locationHeader, '/landed');
          }
          await request.response.close();
        });
    });

    tearDown(() => server.close(force: true));

    test('a 3xx is never followed: it surfaces as an error', () async {
      final dio = _dioFor(
        AppConfig.resolve(
          flavor: AppFlavor.dev,
          apiBaseUrlOverride: 'http://127.0.0.1:${server.port}',
        ),
      );

      await expectLater(
        dio.get<Object?>('/v1/flags'),
        throwsA(
          isA<DioException>().having(
            (e) => e.response?.statusCode,
            'status',
            HttpStatus.found,
          ),
        ),
      );
      expect(served, ['/v1/flags']);
    });
  });

  group('an HTTPS client (staging, prod)', () {
    late Dio dio;
    late RecordingAdapter adapter;

    setUp(() {
      dio = _dioFor(
        AppConfig.resolve(
          flavor: AppFlavor.prod,
          apiBaseUrlOverride: 'https://api.example.test',
        ),
      );
      adapter = RecordingAdapter();
      dio.httpClientAdapter = adapter;
    });

    test('refuses a plain-HTTP request before it leaves the device', () async {
      await expectLater(
        dio.get<Object?>('http://api.example.test/v1/flags'),
        throwsA(
          isA<DioException>().having(
            (e) => e.error,
            'error',
            isA<InsecureRequestError>(),
          ),
        ),
      );
      expect(adapter.requests, isEmpty);
    });

    test('sends HTTPS requests normally', () async {
      final response = await dio.get<Object?>('/v1/flags');

      expect(response.statusCode, HttpStatus.ok);
      expect(
        adapter.requests.single.uri.toString(),
        'https://api.example.test/v1/flags',
      );
    });
  });
}
