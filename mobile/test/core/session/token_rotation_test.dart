import 'package:control_contable/core/network/api_dio.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/recording_adapter.dart';
import '../../helpers/session_fakes.dart';

const _refreshPath = '/v1/auth/refresh';
const _leaked = 'leaked-refresh-token';

void main() {
  late RecordingAdapter adapter;
  late TokenRotation rotate;

  setUp(() {
    adapter = RecordingAdapter();
    final dio = Dio(BaseOptions(baseUrl: 'http://localhost:3001'))
      ..httpClientAdapter = adapter;
    final container = ProviderContainer(
      overrides: [bareDioProvider.overrideWithValue(dio)],
    );
    addTearDown(container.dispose);
    rotate = container.read(tokenRotationProvider);
  });

  test('a well-formed pair is returned', () async {
    adapter.reply(_refreshPath, (
      status: 200,
      body: tokenPair(access: 'a', refresh: 'r').toJson(),
    ));

    final pair = await rotate('refresh-1');

    expect(pair.refreshToken, 'r');
  });

  // M1-16 FU1: the rotation runs inside the auth interceptor; a decode
  // TypeError there escaped `Session._rotateOnce`'s `on ApiFailure`.
  for (final (name, body) in <(String, Object?)>[
    ('access_token: null', {'access_token': null, 'refresh_token': _leaked}),
    (
      'a wrongly typed expiry',
      {
        ...tokenPair(access: 'a', refresh: _leaked).toJson(),
        'access_expires_in': 'soon',
      },
    ),
    ('no body', null),
  ]) {
    test('a malformed refresh answer ($name) is unexpected', () async {
      adapter.reply(_refreshPath, (status: 200, body: body));

      final thrown = await failureOf(() => rotate('refresh-1'));

      expect(thrown, isA<ApiUnexpected>());
      expect('$thrown', isNot(contains(_leaked)));
    });
  }

  test('a malformed refresh answer keeps the session (transient)', () async {
    adapter.reply(_refreshPath, (
      status: 200,
      body: {'access_token': null, 'refresh_token': _leaked},
    ));
    final session = Session(
      store: InMemoryRefreshTokenStore('refresh-1'),
      rotate: rotate,
      storageRetryDelays: const [],
    );
    addTearDown(session.dispose);
    await session.restore();

    expect(await session.refresh(), isNull);
    expect(session.status, SessionStatus.signedIn);
  });
}
