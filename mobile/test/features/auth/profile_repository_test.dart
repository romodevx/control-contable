import 'dart:convert';

import 'package:control_contable/api/generated/api_client.dart';
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

  test('getMe reads the profile', () async {
    final user = await repository.getMe();

    expect(adapter.requests.single.method, 'GET');
    expect(user.email, 'ana.gomez@example.mx');
    expect(user.name, 'Ana Gómez Ruiz');
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

  group('a wrong password is an ApiUnauthorized, whatever the status', () {
    test('401 (the contract today)', () async {
      await expectDeleteFailure((
        status: 401,
        body: <String, Object?>{},
      ), isA<ApiUnauthorized>());
    });

    test('403 invalid_password (a distinct code, mapped by code)', () async {
      await expectDeleteFailure((
        status: 403,
        body: envelope('invalid_password'),
      ), isA<ApiUnauthorized>());
    });

    test('400 invalid_password is still the password', () async {
      await expectDeleteFailure((
        status: 400,
        body: envelope('invalid_password'),
      ), isA<ApiUnauthorized>());
    });
  });

  test('a 403 with any other code is not taken for the password', () async {
    await expectDeleteFailure((
      status: 403,
      body: envelope('forbidden'),
    ), isA<ApiUnexpected>());
  });

  test('getMe never reads a 403 invalid_password as unauthorized', () async {
    adapter.reply('/v1/me', (status: 403, body: envelope('invalid_password')));

    await expectLater(repository.getMe(), throwsA(isA<ApiUnexpected>()));
  });
}
