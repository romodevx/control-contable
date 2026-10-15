import 'dart:convert';

import 'package:control_contable/api/generated/api_client.dart';
import 'package:control_contable/api/generated/models/access_level.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/auth/data/profile_repository.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/recording_adapter.dart';
import '../../helpers/synthetic_password.dart';

const Map<String, Object?> _user = {
  'id': '0199d4a2-7c1e-7b3a-9f10-2c5e8a1b4d01',
  'email': 'ana.gomez@example.mx',
  'name': 'Ana Gómez Ruiz',
  'created_at': '2026-10-13T16:00:00Z',
  'access': {
    'access': 'trial',
    'trial_ends_at': '2026-10-20T16:00:00Z',
    'trial_days_left': 3,
    'incomes_remaining': 2,
    'expenses_remaining': 5,
  },
};

void main() {
  late RecordingAdapter adapter;
  late ApiProfileRepository repository;

  setUp(() {
    adapter = RecordingAdapter(replies: {'/v1/me': (status: 200, body: _user)});
    final dio = Dio(BaseOptions(baseUrl: 'http://localhost:3001'))
      ..httpClientAdapter = adapter;
    repository = ApiProfileRepository(ApiClient(dio).auth);
  });

  Map<String, Object?> lastBody() =>
      jsonDecode(jsonEncode(adapter.requests.last.data))
          as Map<String, Object?>;

  test('getMe reads the profile and the access state', () async {
    final me = await repository.getMe();

    expect(adapter.requests.single.method, 'GET');
    expect(me.email, 'ana.gomez@example.mx');
    expect(me.name, 'Ana Gómez Ruiz');
    expect(me.access.access, AccessLevel.trial);
    expect(me.access.trialDaysLeft, 3);
    expect(me.access.incomesRemaining, 2);
    expect(me.access.expensesRemaining, 5);
  });

  test('updateName patches only the name', () async {
    await repository.updateName('Ana Gómez de la Torre');

    expect(adapter.requests.single.method, 'PATCH');
    expect(lastBody(), {'name': 'Ana Gómez de la Torre'});
  });

  test('deleteAccount sends the current password and accepts 204', () async {
    adapter.reply('/v1/me', (status: 204, body: null));
    final password = syntheticPassword();

    await repository.deleteAccount(password);

    expect(adapter.requests.single.method, 'DELETE');
    expect(lastBody(), {'password': password});
  });

  Map<String, Object?> envelope(String code) => {
    'error': {'code': code, 'message': 'Refused.', 'request_id': 'req_1'},
  };

  Future<void> expectDeleteFailure(
    ({int status, Object? body}) reply,
    Matcher failure,
  ) async {
    adapter.reply('/v1/me', reply);
    await expectLater(
      repository.deleteAccount(syntheticPassword()),
      throwsA(failure),
    );
  }

  test(
    'a wrong password (403 invalid_password) is an ApiInvalidPassword',
    () async {
      await expectDeleteFailure((
        status: 403,
        body: envelope('invalid_password'),
      ), isA<ApiInvalidPassword>());
    },
  );

  test('a 401 on delete is the bearer, never the password', () async {
    await expectDeleteFailure((
      status: 401,
      body: envelope('unauthorized'),
    ), isA<ApiUnauthorized>());
  });

  test('too many tries are an ApiRateLimited', () async {
    await expectDeleteFailure((
      status: 429,
      body: envelope('rate_limited'),
    ), isA<ApiRateLimited>());
  });
  // M1-16 FU1: a profile that does not match the contract is the generic
  // ApiUnexpected, never a raw TypeError carrying the payload.
  group('a malformed profile is unexpected', () {
    const leaked = 'leaked@example.mx';
    for (final (name, body) in <(String, Object?)>[
      ('a null required field', {..._user, 'email': null}),
      ('a wrongly typed field', {..._user, 'id': 7, 'email': leaked}),
      ('an unreadable date', {..._user, 'created_at': leaked}),
      ('no body at all', null),
    ]) {
      test('getMe: $name', () async {
        adapter.reply('/v1/me', (status: 200, body: body));

        final thrown = await failureOf(repository.getMe);

        expect(thrown, isA<ApiUnexpected>());
        expect('$thrown', isNot(contains(leaked)));
      });

      test('updateName: $name', () async {
        adapter.reply('/v1/me', (status: 200, body: body));

        final thrown = await failureOf(() => repository.updateName('Ana'));

        expect(thrown, isA<ApiUnexpected>());
        expect('$thrown', isNot(contains(leaked)));
      });
    }
  });
}
