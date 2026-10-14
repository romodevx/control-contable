import 'dart:async';

import 'package:control_contable/api/generated/models/token_pair.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/session_fakes.dart';

/// A scripted `refreshSession`: records the refresh tokens it was given and
/// answers with [next] (or throws [error]); a [gate] holds it open.
class _Rotation {
  final List<String> received = [];
  TokenPair next = tokenPair(access: 'access-2', refresh: 'refresh-2');
  Exception? error;
  Completer<void>? gate;

  Future<TokenPair> call(String refreshToken) async {
    received.add(refreshToken);
    await gate?.future;
    if (error case final error?) throw error;
    return next;
  }
}

void main() {
  late InMemoryRefreshTokenStore store;
  late _Rotation rotation;
  late Session session;

  setUp(() {
    store = InMemoryRefreshTokenStore();
    rotation = _Rotation();
    session = Session(store: store, rotate: rotation.call);
    addTearDown(session.dispose);
  });

  group('restore', () {
    test('starts restoring', () {
      expect(session.status, SessionStatus.restoring);
    });

    test('a stored refresh token means signed in', () async {
      store.token = 'refresh-1';

      await session.restore();

      expect(session.status, SessionStatus.signedIn);
      expect(session.accessToken, isNull, reason: 'memory only');
    });

    test('no stored token means signed out', () async {
      await session.restore();

      expect(session.status, SessionStatus.signedOut);
    });

    test('an unreadable store means signed out', () async {
      store.readError = PlatformException(code: 'keychain');

      await session.restore();

      expect(session.status, SessionStatus.signedOut);
    });

    test('notifies its listeners', () async {
      var notified = 0;
      session.addListener(() => notified++);

      await session.restore();

      expect(notified, 1);
    });
  });

  test(
    'start keeps the refresh token in the store, the access in memory',
    () async {
      await session.start(tokenPair(access: 'access-1', refresh: 'refresh-1'));

      expect(store.token, 'refresh-1');
      expect(session.accessToken, 'access-1');
      expect(session.status, SessionStatus.signedIn);
    },
  );

  group('refresh', () {
    setUp(
      () => session.start(tokenPair(access: 'access-1', refresh: 'refresh-1')),
    );

    test('rotates: sends the stored token, keeps the new one', () async {
      final access = await session.refresh();

      expect(access, 'access-2');
      expect(rotation.received, ['refresh-1']);
      expect(store.token, 'refresh-2', reason: 'the old one is dead now');
      expect(session.accessToken, 'access-2');
    });

    test('the next rotation sends the newest token', () async {
      await session.refresh();
      rotation.next = tokenPair(access: 'access-3', refresh: 'refresh-3');

      await session.refresh();

      expect(rotation.received, ['refresh-1', 'refresh-2']);
      expect(store.token, 'refresh-3');
    });

    test('concurrent callers share one rotation (single flight)', () async {
      rotation.gate = Completer<void>();

      final waiting = [for (var i = 0; i < 5; i++) session.refresh()];
      rotation.gate!.complete();
      final tokens = await Future.wait(waiting);

      expect(rotation.received, hasLength(1));
      expect(tokens, everyElement('access-2'));
    });

    test('a later refresh starts a new rotation', () async {
      await session.refresh();
      await session.refresh();

      expect(rotation.received, hasLength(2));
    });

    for (final rejection in <ApiFailure>[
      const ApiUnauthorized(),
      const ApiValidationFailure([]),
    ]) {
      test(
        'a rejected token (${rejection.runtimeType}) ends the session',
        () async {
          rotation.error = rejection;

          final access = await session.refresh();

          expect(access, isNull);
          expect(session.status, SessionStatus.signedOut);
          expect(session.accessToken, isNull);
          expect(store.token, isNull);
        },
      );
    }

    for (final transient in <ApiFailure>[
      const ApiUnreachable(),
      const ApiUnexpected(),
      const ApiRateLimited(null),
    ]) {
      test('a transient failure (${transient.runtimeType}) keeps it', () async {
        rotation.error = transient;

        final access = await session.refresh();

        expect(access, isNull);
        expect(session.status, SessionStatus.signedIn);
        expect(store.token, 'refresh-1');
      });
    }

    test('with no stored token ends the session without a call', () async {
      store.token = null;

      expect(await session.refresh(), isNull);
      expect(rotation.received, isEmpty);
      expect(session.status, SessionStatus.signedOut);
    });

    test('ending during a rotation never brings the session back', () async {
      rotation.gate = Completer<void>();
      final refreshing = session.refresh();

      await session.end();
      rotation.gate!.complete();

      expect(await refreshing, isNull);
      expect(store.token, isNull);
      expect(session.status, SessionStatus.signedOut);
    });
  });

  test('end forgets both tokens and signs out', () async {
    await session.start(tokenPair(access: 'access-1', refresh: 'refresh-1'));

    await session.end();

    expect(store.token, isNull);
    expect(session.accessToken, isNull);
    expect(session.status, SessionStatus.signedOut);
  });

  test('a sign-in during restore wins over the stored answer', () async {
    final restoring = session.restore();
    await session.start(tokenPair(access: 'a', refresh: 'r'));
    await restoring;

    expect(session.status, SessionStatus.signedIn);
  });
}
