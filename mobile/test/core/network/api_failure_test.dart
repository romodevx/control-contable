import 'package:control_contable/api/generated/models/token_pair.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:json_annotation/json_annotation.dart';

import '../../helpers/recording_adapter.dart';

final _now = DateTime.utc(2026, 10, 13, 12);
final _request = RequestOptions(path: '/v1/auth/register');

DioException _answer(
  int status, {
  Object? body,
  Map<String, List<String>> headers = const {},
}) => DioException.badResponse(
  statusCode: status,
  requestOptions: _request,
  response: Response<Object?>(
    requestOptions: _request,
    statusCode: status,
    data: body,
    headers: Headers.fromMap(headers),
  ),
);

Map<String, Object?> _envelope(String code, [List<Object?>? details]) => {
  'error': {
    'code': code,
    'message': 'One or more fields are invalid.',
    'request_id': 'req_1',
    'details': ?details,
  },
};

ApiFailure _failureOf(DioException exception) =>
    apiFailureFrom(exception, now: () => _now);

Map<String, Object?> _tokens(String access) => {
  'access_token': access,
  'access_expires_in': 900,
  'refresh_token': 'refresh-1',
  'refresh_expires_in': 2592000,
  'token_type': 'Bearer',
};

Map<String, Object?>? _noBody() => null;

Future<Object?> _guarded(Future<Object?> Function() call) =>
    failureOf(() => guardApi(call));

void main() {
  group('apiFailureFrom', () {
    group('invalid_password is read by its code, not its status', () {
      for (final status in [403, 400, 401]) {
        test('$status invalid_password is an ApiInvalidPassword', () {
          final failure = _failureOf(
            _answer(status, body: _envelope('invalid_password')),
          );

          expect(failure, isA<ApiInvalidPassword>());
        });
      }

      test('a 403 with another code is unexpected', () {
        expect(
          _failureOf(_answer(403, body: _envelope('conflict'))),
          isA<ApiUnexpected>(),
        );
      });

      test('a 401 without that code is the bearer', () {
        expect(
          _failureOf(_answer(401, body: _envelope('unauthorized'))),
          isA<ApiUnauthorized>(),
        );
      });
    });

    test('400 carries the field details of the error envelope', () {
      final failure = _failureOf(
        _answer(
          400,
          body: _envelope('validation_failed', [
            {
              'field': 'password',
              'code': 'password_policy',
              'message': 'Use at least 12 characters.',
            },
          ]),
        ),
      );

      expect(failure, isA<ApiValidationFailure>());
      final details = (failure as ApiValidationFailure).details;
      expect(details.single.field, 'password');
      expect(details.single.code, 'password_policy');
    });

    test('400 without a readable envelope has no details', () {
      final failure = _failureOf(_answer(400, body: 'oops'));

      expect(
        failure,
        isA<ApiValidationFailure>().having(
          (f) => f.details,
          'details',
          isEmpty,
        ),
      );
    });

    test('401 is unauthorized', () {
      expect(
        _failureOf(_answer(401, body: _envelope('unauthorized'))),
        isA<ApiUnauthorized>(),
      );
    });

    test('404 is not found (a hidden module answers it too)', () {
      expect(_failureOf(_answer(404)), isA<ApiNotFound>());
    });

    test('429 carries Retry-After in seconds', () {
      final failure = _failureOf(
        _answer(
          429,
          headers: {
            'retry-after': ['42'],
          },
        ),
      );

      expect(
        failure,
        isA<ApiRateLimited>().having(
          (f) => f.retryAfter,
          'retryAfter',
          const Duration(seconds: 42),
        ),
      );
    });

    test('429 without Retry-After has no wait', () {
      expect((_failureOf(_answer(429)) as ApiRateLimited).retryAfter, isNull);
    });

    test('5xx and unknown statuses are unexpected', () {
      expect(_failureOf(_answer(500)), isA<ApiUnexpected>());
      expect(_failureOf(_answer(418)), isA<ApiUnexpected>());
    });

    test('no answer at all (offline, timeout) is unreachable', () {
      for (final type in [
        DioExceptionType.connectionError,
        DioExceptionType.connectionTimeout,
        DioExceptionType.receiveTimeout,
        DioExceptionType.sendTimeout,
      ]) {
        expect(
          _failureOf(DioException(requestOptions: _request, type: type)),
          isA<ApiUnreachable>(),
          reason: '$type',
        );
      }
    });
  });

  group('parseRetryAfter', () {
    test('reads delta-seconds', () {
      expect(parseRetryAfter('60', now: _now), const Duration(seconds: 60));
    });

    test('reads an HTTP-date relative to now', () {
      expect(
        parseRetryAfter('Tue, 13 Oct 2026 12:01:30 GMT', now: _now),
        const Duration(seconds: 90),
      );
    });

    test('a past date, a negative or a garbage value is no wait', () {
      expect(parseRetryAfter('Tue, 13 Oct 2026 11:00:00 GMT', now: _now), null);
      expect(parseRetryAfter('-5', now: _now), isNull);
      expect(parseRetryAfter('soon', now: _now), isNull);
      expect(parseRetryAfter(null, now: _now), isNull);
    });
  });

  group('guardApi', () {
    test('passes a result through', () async {
      expect(await guardApi(() async => 7), 7);
    });

    test('turns a DioException into an ApiFailure', () async {
      await expectLater(
        guardApi<void>(() => Future.error(_answer(401))),
        throwsA(isA<ApiUnauthorized>()),
      );
    });

    test('a body the client cannot decode is unexpected', () async {
      await expectLater(
        guardApi<void>(() => Future.error(const FormatException('bad json'))),
        throwsA(isA<ApiUnexpected>()),
      );
    });

    // M1-16 FU1: a generated decoder that casts a null or wrongly typed
    // field throws a TypeError, an Error that `on Exception` let through.
    group('a payload that breaks the generated decoder is unexpected', () {
      const leakedToken = 'secret-access-token';
      final payloads = <String, Map<String, Object?>>{
        'a null required field': {
          ..._tokens(leakedToken),
          'refresh_token': null,
        },
        'a wrongly typed field': {
          ..._tokens(leakedToken),
          'access_expires_in': 'soon',
        },
        'a missing field': {'access_token': leakedToken},
      };
      for (final MapEntry(key: name, value: payload) in payloads.entries) {
        test(name, () async {
          final thrown = await _guarded(
            () async => TokenPair.fromJson(payload),
          );

          expect(thrown, isA<ApiUnexpected>());
          expect('$thrown', isNot(contains(leakedToken)));
        });
      }

      test('a null body where an object was expected', () async {
        expect(
          await _guarded(() async => TokenPair.fromJson(_noBody()!)),
          isA<ApiUnexpected>(),
        );
      });

      test('a CheckedFromJsonException', () async {
        expect(
          await _guarded(
            () => Future.error(
              CheckedFromJsonException({}, 'access_token', 'TokenPair', null),
            ),
          ),
          isA<ApiUnexpected>(),
        );
      });
    });

    test('a programming error is not disguised as an API failure', () async {
      await expectLater(
        guardApi<void>(() => Future.error(StateError('bug'))),
        throwsA(isA<StateError>()),
      );
    });
  });
}
