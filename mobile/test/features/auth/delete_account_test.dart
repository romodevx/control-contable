import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/auth/application/delete_account.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/session_fakes.dart';
import '../../helpers/synthetic_password.dart';
import 'auth_fakes.dart';

void main() {
  late FakeProfileRepository repository;
  late InMemoryRefreshTokenStore store;
  late Session session;
  late DeleteAccount deleteAccount;

  setUp(() async {
    repository = FakeProfileRepository();
    store = InMemoryRefreshTokenStore();
    session = Session(
      store: store,
      rotate: (_) async => tokenPair(access: 'a2', refresh: 'r2'),
    );
    addTearDown(session.dispose);
    await session.start(tokenPair(access: 'a1', refresh: 'r1'));
    deleteAccount = DeleteAccount(repository, session);
  });

  test('deletes in one call (the interceptor refreshes a stale bearer), '
      'then forgets the session', () async {
    await deleteAccount(syntheticPassword());

    expect(repository.calls, ['deleteAccount']);
    expect(repository.deletePasswords, [syntheticPassword()]);
    expect(session.status, SessionStatus.signedOut);
    expect(store.token, isNull);
  });

  test('a wrong password deletes nothing and keeps the session', () async {
    repository.deleteError = const ApiInvalidPassword();

    await expectLater(
      deleteAccount(syntheticPassword()),
      throwsA(isA<ApiInvalidPassword>()),
    );

    expect(session.status, SessionStatus.signedIn);
    expect(store.token, 'r1');
  });

  test('when the API cannot be reached the session is kept', () async {
    repository.deleteError = const ApiUnreachable();

    await expectLater(
      deleteAccount(syntheticPassword()),
      throwsA(isA<ApiUnreachable>()),
    );

    expect(session.status, SessionStatus.signedIn);
  });
}
