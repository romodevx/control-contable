import 'package:control_contable/core/access/read_only_access.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/session_fakes.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  late ProviderContainer container;
  late Session session;

  setUp(() {
    session = Session(
      store: InMemoryRefreshTokenStore(),
      rotate: (_) async => tokenPair(access: 'a2', refresh: 'r2'),
    );
    container = ProviderContainer(
      overrides: [sessionProvider.overrideWithValue(session)],
    );
  });
  tearDown(() {
    container.dispose();
    session.dispose();
  });

  bool known() => container.read(readOnlyAccessProvider);
  void mark() => container.read(readOnlyAccessProvider.notifier).markReadOnly();

  test('nothing is known until the server answers 402', () {
    expect(known(), isFalse);
    mark();
    expect(known(), isTrue);
  });

  test('a sign-out forgets it: the next account starts unknown', () async {
    await session.start(tokenPair(access: 'a1', refresh: 'r1'));
    mark();

    await session.end();

    expect(known(), isFalse);
  });

  test('returning to the app forgets it: a plan may have been bought in '
      'the store meanwhile', () {
    mark();

    WidgetsBinding.instance
      ..handleAppLifecycleStateChanged(AppLifecycleState.inactive)
      ..handleAppLifecycleStateChanged(AppLifecycleState.resumed);

    expect(known(), isFalse);
  });
}
