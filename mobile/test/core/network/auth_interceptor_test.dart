import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:control_contable/api/generated/models/token_pair.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/network/auth_interceptor.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/session_fakes.dart';

const _protected = '/v1/me';
const _fresh = 'access-2';
final Uri _base = Uri.parse('https://api.test');

/// An API that accepts only the bearer in [validAccess] and answers 401 to
/// anything else (every path, auth endpoints included). Counts its 401s so
/// a test can hold the refresh until every request has been refused.
class _Api implements HttpClientAdapter {
  String validAccess = _fresh;
  final List<RequestOptions> requests = [];
  int refusals = 0;
  void Function()? onRefusal;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    requests.add(options);
    final authorized =
        options.headers[HttpHeaders.authorizationHeader] ==
        'Bearer $validAccess';
    if (!authorized) {
      refusals++;
      onRefusal?.call();
    }
    return ResponseBody.fromString(
      jsonEncode(<String, Object?>{}),
      authorized ? HttpStatus.ok : HttpStatus.unauthorized,
      headers: {
        Headers.contentTypeHeader: [ContentType.json.mimeType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}

class _Rotation {
  int calls = 0;
  Exception? error;
  Completer<void>? gate;

  Future<TokenPair> call(String refreshToken) async {
    calls++;
    await gate?.future;
    if (error case final error?) throw error;
    return tokenPair(access: _fresh, refresh: 'refresh-2');
  }
}

void main() {
  late _Api api;
  late _Rotation rotation;
  late InMemoryRefreshTokenStore store;
  late Session session;
  late Dio dio;

  setUp(() async {
    api = _Api();
    rotation = _Rotation();
    store = InMemoryRefreshTokenStore();
    session = Session(store: store, rotate: rotation.call);
    addTearDown(session.dispose);
    dio = Dio(BaseOptions(baseUrl: 'https://api.test'))
      ..httpClientAdapter = api;
    dio.interceptors.add(
      AuthInterceptor(session: session, dio: dio, apiBase: _base),
    );
    await session.start(tokenPair(access: 'access-1', refresh: 'refresh-1'));
  });

  String? bearerOf(RequestOptions request) =>
      request.headers[HttpHeaders.authorizationHeader] as String?;

  test('attaches the in-memory access token as a bearer', () async {
    api.validAccess = 'access-1';

    await dio.get<void>(_protected);

    expect(bearerOf(api.requests.single), 'Bearer access-1');
  });

  for (final path in [
    '/v1/auth/login',
    '/v1/auth/refresh',
    '/v1/auth/register',
    '/v1/auth/verify-email',
    '/v1/auth/verify-email/resend',
  ]) {
    test('sends no bearer to $path and never refreshes on its 401', () async {
      await expectLater(dio.post<void>(path), throwsA(isA<DioException>()));

      expect(bearerOf(api.requests.single), isNull);
      expect(rotation.calls, 0);
    });
  }

  test('logout carries the bearer but its 401 never refreshes', () async {
    await expectLater(
      dio.post<void>('/v1/auth/logout'),
      throwsA(isA<DioException>()),
    );

    expect(bearerOf(api.requests.single), 'Bearer access-1');
    expect(rotation.calls, 0);
  });

  test('a 401 refreshes once and retries with the new token', () async {
    final response = await dio.get<void>(_protected);

    expect(response.statusCode, HttpStatus.ok);
    expect(rotation.calls, 1);
    expect(api.requests.map(bearerOf), ['Bearer access-1', 'Bearer $_fresh']);
    expect(store.token, 'refresh-2', reason: 'rotation-safe');
  });

  test('5 parallel 401s cause exactly 1 refresh, and all succeed', () async {
    const parallel = 5;
    rotation.gate = Completer<void>();
    api.onRefusal = () {
      if (api.refusals == parallel) rotation.gate!.complete();
    };

    final responses = await Future.wait([
      for (var i = 0; i < parallel; i++) dio.get<void>(_protected),
    ]);

    expect(rotation.calls, 1);
    expect(responses.map((r) => r.statusCode), everyElement(HttpStatus.ok));
    expect(api.requests, hasLength(parallel * 2));
  });

  test(
    'a 401 for a token already replaced retries without refreshing',
    () async {
      // The request left with access-1; a refresh finished meanwhile.
      await session.refresh();
      expect(rotation.calls, 1);
      final stale = RequestOptions(
        path: _protected,
        baseUrl: 'https://api.test',
        headers: {HttpHeaders.authorizationHeader: 'Bearer access-1'},
      );
      final handler = _RecordingErrorHandler();

      await AuthInterceptor(session: session, dio: dio, apiBase: _base).onError(
        DioException(
          requestOptions: stale,
          response: Response(
            requestOptions: stale,
            statusCode: HttpStatus.unauthorized,
          ),
        ),
        handler,
      );

      expect(rotation.calls, 1);
      expect(handler.resolved?.statusCode, HttpStatus.ok);
    },
  );

  test('a retry refused again is not refreshed a second time', () async {
    api.validAccess = 'never';

    await expectLater(
      dio.get<void>(_protected),
      throwsA(
        isA<DioException>().having(
          (e) => e.response?.statusCode,
          'status',
          HttpStatus.unauthorized,
        ),
      ),
    );

    expect(rotation.calls, 1);
    expect(api.requests, hasLength(2));
  });

  test('a rejected refresh ends the session and surfaces the 401', () async {
    rotation.error = const ApiUnauthorized();

    await expectLater(dio.get<void>(_protected), throwsA(isA<DioException>()));

    expect(session.status, SessionStatus.signedOut);
    expect(store.token, isNull);
    expect(api.requests, hasLength(1), reason: 'no retry');
  });

  test('a refresh that cannot reach the server keeps the session', () async {
    rotation.error = const ApiUnreachable();

    await expectLater(dio.get<void>(_protected), throwsA(isA<DioException>()));

    expect(session.status, SessionStatus.signedIn);
    expect(store.token, 'refresh-1');
  });

  test('signed out: a 401 passes through untouched', () async {
    await session.end();

    await expectLater(dio.get<void>(_protected), throwsA(isA<DioException>()));

    expect(bearerOf(api.requests.single), isNull);
    expect(rotation.calls, 0);
  });

  test(
    'after a cold start, the first call refreshes before it leaves',
    () async {
      final restored = Session(store: store, rotate: rotation.call);
      addTearDown(restored.dispose);
      await restored.restore();
      dio.interceptors
        ..clear()
        ..add(AuthInterceptor(session: restored, dio: dio, apiBase: _base));

      await dio.get<void>(_protected);

      expect(rotation.calls, 1);
      expect(api.requests.map(bearerOf), ['Bearer $_fresh'], reason: 'no 401');
    },
  );

  group(
    'only the configured API host gets the bearer (PR #42 follow-up a)',
    () {
      for (final foreign in [
        'https://evil.test/v1/me',
        'http://api.test/v1/me',
        'https://api.test:8443/v1/me',
        'https://api.test.evil.test/v1/me',
      ]) {
        test('no bearer and no refresh for $foreign', () async {
          await expectLater(
            dio.get<void>(foreign),
            throwsA(isA<DioException>()),
          );

          expect(bearerOf(api.requests.single), isNull);
          expect(rotation.calls, 0);
        });
      }

      test('a cold start never refreshes for a foreign host', () async {
        final restored = Session(store: store, rotate: rotation.call);
        addTearDown(restored.dispose);
        await restored.restore();
        dio.interceptors
          ..clear()
          ..add(AuthInterceptor(session: restored, dio: dio, apiBase: _base));

        await expectLater(
          dio.get<void>('https://evil.test/v1/me'),
          throwsA(isA<DioException>()),
        );

        expect(rotation.calls, 0);
        expect(bearerOf(api.requests.single), isNull);
      });

      test('the host is compared without regard to case', () async {
        api.validAccess = 'access-1';

        await dio.get<void>('https://API.test/v1/me');

        expect(bearerOf(api.requests.single), 'Bearer access-1');
      });
    },
  );

  test('a 401 to DELETE /v1/me (wrong current password) is answered as is: '
      'no refresh, no second try', () async {
    await expectLater(
      dio.delete<void>(_protected, data: {'password': 'x'}),
      throwsA(
        isA<DioException>().having(
          (e) => e.response?.statusCode,
          'status',
          HttpStatus.unauthorized,
        ),
      ),
    );

    expect(rotation.calls, 0);
    expect(api.requests, hasLength(1));
    expect(session.status, SessionStatus.signedIn);
  });

  test('other errors pass through without a refresh', () async {
    api.validAccess = 'access-1';
    dio.httpClientAdapter = _Failing();

    await expectLater(dio.get<void>(_protected), throwsA(isA<DioException>()));

    expect(rotation.calls, 0);
  });
}

class _Failing implements HttpClientAdapter {
  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async => ResponseBody.fromString('{}', HttpStatus.internalServerError);

  @override
  void close({bool force = false}) {}
}

class _RecordingErrorHandler extends ErrorInterceptorHandler {
  Response<dynamic>? resolved;
  DioException? passed;

  @override
  void resolve(Response<dynamic> response) => resolved = response;

  @override
  void next(DioException error) => passed = error;
}
