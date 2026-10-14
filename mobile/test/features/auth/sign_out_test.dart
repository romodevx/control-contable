import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/auth/application/sign_out.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/recording_adapter.dart';
import '../../helpers/recording_error_reporter.dart';
import '../../helpers/session_fakes.dart';
import 'auth_fakes.dart';

const _stored = 'refresh-stored';

void main() {
  late InMemoryRefreshTokenStore store;
  late FakeTokenRotation rotation;
  late FakeAuthRepository repository;
  late Session session;
  late SignOut signOut;
  late RecordingErrorReporter reporter;

  setUp(() async {
    reporter = RecordingErrorReporter();
    store = InMemoryRefreshTokenStore(_stored);
    rotation = FakeTokenRotation();
    repository = FakeAuthRepository();
    session = Session(
      store: store,
      rotate: rotation.call,
      storageRetryDelays: const [],
      reporter: reporter,
    );
    addTearDown(session.dispose);
    await session.restore();
    signOut = SignOut(repository, session, reporter);
  });

  void expectSignedOutLocally() {
    expect(session.status, SessionStatus.signedOut);
    expect(session.accessToken, isNull);
  }

  test('revokes the stored refresh token, then wipes it', () async {
    await session.refresh();

    await signOut();

    expect(repository.loggedOut, ['refresh-rotated']);
    expect(store.token, isNull);
    expectSignedOutLocally();
    expect(reporter.reports, isEmpty);
  });

  // M1-16 FU2: a store read that fails while signing out escaped
  // `SignOut.call` (only `ApiFailure` was caught) from an unawaited tap.
  group('an unreadable secure store', () {
    setUp(
      () => store.readError = PlatformException(code: 'read', message: _stored),
    );

    test('still signs out locally and never throws', () async {
      await session.refresh(); // fails to read: no access token

      final thrown = await failureOf(signOut.call);

      expect(thrown, isNull);
      expect(repository.loggedOut, isEmpty, reason: 'no token to revoke');
      expect(reporter.reports.map((r) => r.operation), [
        'session.readRefreshToken',
      ]);
      expect(store.token, isNull, reason: 'the delete still ran');
      expectSignedOutLocally();
    });

    test('with an access token in memory, still signs out', () async {
      await session.start(tokenPair(access: 'access-1', refresh: _stored));
      store.readError = PlatformException(code: 'read', message: _stored);

      final thrown = await failureOf(signOut.call);

      expect(thrown, isNull);
      expectSignedOutLocally();
    });

    test('and an undeletable token is reported, not thrown', () async {
      store
        ..failingDeletes = storageAttempts
        ..wipeError = PlatformException(code: 'wipe', message: _stored);

      final thrown = await failureOf(signOut.call);

      expect(thrown, isNull);
      expect(session.tokenLeftOnDevice, isTrue);
      expectSignedOutLocally();
    });
  });

  for (final failure in <Exception>[
    const ApiUnreachable(),
    const ApiUnexpected(),
    PlatformException(code: 'unexpected', message: _stored),
  ]) {
    test('a failed revocation (${failure.runtimeType}) still signs out '
        'and never throws', () async {
      await session.refresh();
      repository.logoutError = failure;

      final thrown = await failureOf(signOut.call);

      expect(thrown, isNull);
      expect(store.token, isNull);
      expectSignedOutLocally();
      expect(reporter.reports.single.operation, 'signOut.revoke');
      expect(reporter.reports.single.error, same(failure));
    });
  }
}
