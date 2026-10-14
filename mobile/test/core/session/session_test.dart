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
    session = Session(
      store: store,
      rotate: rotation.call,
      storageRetryDelays: _noWait,
    );
    addTearDown(session.dispose);
  });

  test('retries a failed store call twice before giving up', () {
    expect(storageAttempts, 3);
    expect(defaultStorageRetryDelays, hasLength(storageAttempts - 1));
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

    // M1-15 QA D1: the first cold start of a debug build showed the login
    // screen although the refresh token was still on disk. The Keystore
    // read took longer than the restore timeout, and restore() took the
    // timeout for "no session" for good. The splash stays bounded, but the
    // read's late answer must still restore the session.
    testWidgets('a read slower than the timeout still restores the session '
        'once it answers', (tester) async {
      store
        ..token = 'refresh-1'
        ..readGate = Completer<void>();

      unawaited(session.restore());
      await tester.pump(sessionRestoreTimeout);
      expect(session.status, SessionStatus.signedOut, reason: 'splash bounded');

      store.readGate!.complete();
      await tester.pump();
      expect(session.status, SessionStatus.signedIn);
    });

    testWidgets('a late answer never revives a session signed in and out '
        'meanwhile', (tester) async {
      store
        ..token = 'refresh-old'
        ..readGate = Completer<void>();

      unawaited(session.restore());
      await tester.pump(sessionRestoreTimeout);
      await session.start(tokenPair(access: 'access-new', refresh: 'new'));
      await session.end();
      store.readGate!.complete();
      await tester.pump();

      expect(session.status, SessionStatus.signedOut);
      expect(store.token, isNull);
    });

    testWidgets('a late empty answer stays signed out, told once', (
      tester,
    ) async {
      store.readGate = Completer<void>();
      var notified = 0;
      session.addListener(() => notified++);

      unawaited(session.restore());
      await tester.pump(sessionRestoreTimeout);
      store.readGate!.complete();
      await tester.pump();

      expect(session.status, SessionStatus.signedOut);
      expect(notified, 1);
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

  group('a failed save after a rotation (PR #42 follow-up b)', () {
    setUp(() async {
      await session.start(tokenPair(access: 'access-1', refresh: 'refresh-1'));
      store.written.clear();
    });

    test('retries the save until the store takes it', () async {
      store.failingWrites = 1;

      expect(await session.refresh(), 'access-2');

      expect(store.token, 'refresh-2');
      expect(store.written, ['refresh-2', 'refresh-2']);
    });

    test('keeps the session in memory and surfaces no secret', () async {
      store.failingWrites = _always;

      // Never throws: the PlatformException echoes the token it failed on.
      expect(await session.refresh(), 'access-2');

      expect(session.status, SessionStatus.signedIn);
      expect(session.accessToken, 'access-2');
      expect(store.written, hasLength(storageAttempts));
    });

    test('the next rotation sends the unsaved token, never the dead '
        'one still on disk (that would be reuse)', () async {
      store.failingWrites = _always;
      await session.refresh();
      rotation.next = tokenPair(access: 'access-3', refresh: 'refresh-3');

      await session.refresh();

      expect(store.token, 'refresh-1', reason: 'the disk still lags');
      expect(rotation.received, ['refresh-1', 'refresh-2']);
    });

    test('once the store works again, the newest token is saved', () async {
      store.failingWrites = storageAttempts;
      await session.refresh();
      rotation.next = tokenPair(access: 'access-3', refresh: 'refresh-3');

      await session.refresh();

      expect(store.token, 'refresh-3');
      expect(await session.readRefreshToken(), 'refresh-3');
    });

    test('logout reads the unsaved token', () async {
      store.failingWrites = _always;
      await session.refresh();

      expect(await session.readRefreshToken(), 'refresh-2');
    });

    test('a failed save at sign-in still signs in, in memory', () async {
      await session.end();
      store.failingWrites = _always;

      await session.start(tokenPair(access: 'access-9', refresh: 'refresh-9'));

      expect(session.status, SessionStatus.signedIn);
      expect(await session.readRefreshToken(), 'refresh-9');
    });
  });

  // M1-16 FU2: the store's exception (which may echo the token) must not
  // escape to the sign-out that asked.
  test('readRefreshToken is null when the store cannot be read', () async {
    store
      ..token = 'refresh-1'
      ..readError = PlatformException(code: 'read', message: 'refresh-1');

    expect(await session.readRefreshToken(), isNull);
  });

  group('a failed delete at sign-out (PR #42 follow-up c)', () {
    setUp(
      () => session.start(tokenPair(access: 'access-1', refresh: 'refresh-1')),
    );

    test('retries the delete', () async {
      store.failingDeletes = 1;

      await session.end();

      expect(store.token, isNull);
      expect(store.wipes, 0);
      expect(session.tokenLeftOnDevice, isFalse);
    });

    test('wipes the whole namespace when the delete keeps failing', () async {
      store.failingDeletes = _always;

      await session.end();

      expect(store.wipes, 1);
      expect(store.token, isNull);
      expect(session.tokenLeftOnDevice, isFalse);
    });

    test('never fails silently: reports a token left on the device', () async {
      store
        ..failingDeletes = _always
        ..wipeError = PlatformException(code: 'wipe');
      var notified = 0;
      session.addListener(() => notified++);

      await session.end();

      expect(session.status, SessionStatus.signedOut);
      expect(session.accessToken, isNull);
      expect(session.tokenLeftOnDevice, isTrue);
      expect(notified, 2, reason: 'signed out, then the warning');
    });

    test('the next sign-in retries the wipe and clears the warning', () async {
      store
        ..failingDeletes = _always
        ..wipeError = PlatformException(code: 'wipe');
      await session.end();
      store
        ..failingDeletes = 0
        ..wipeError = null;

      await session.start(tokenPair(access: 'a', refresh: 'r'));

      expect(session.tokenLeftOnDevice, isFalse);
      expect(store.token, 'r');
    });
  });
}

/// Retries at once: the waits are not what these tests are about.
const List<Duration> _noWait = [Duration.zero, Duration.zero];

/// Enough failures to outlast every retry.
const int _always = 1 << 20;
