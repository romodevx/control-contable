import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/taxpayers/application/active_taxpayer.dart';
import 'package:control_contable/features/taxpayers/application/taxpayer_list.dart';
import 'package:control_contable/features/taxpayers/data/active_taxpayer_store.dart';
import 'package:control_contable/features/taxpayers/data/taxpayers_repository.dart';
import 'package:control_contable/features/taxpayers/domain/masked_rfc.dart';

import 'dart:async';

import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/session_fakes.dart';
import 'taxpayers_fakes.dart';

final Taxpayer _first = taxpayer();
final Taxpayer _second = taxpayer(
  id: secondTaxpayerId,
  displayName: 'Taller',
  createdAt: DateTime.utc(2026, 10, 14),
);
final Taxpayer _archived = taxpayer(
  id: archivedTaxpayerId,
  displayName: 'Viejo',
  archivedAt: DateTime.utc(2026, 10, 12),
  createdAt: DateTime.utc(2026, 1, 2),
);

/// A repository whose next list answer is held by [held] (then plain).
class _HeldListRepository extends FakeTaxpayersRepository {
  Completer<List<Taxpayer>>? held;

  @override
  Future<List<Taxpayer>> listTaxpayers() {
    final answer = held;
    if (answer == null) return super.listTaxpayers();
    held = null;
    return answer.future;
  }
}

void main() {
  final binding = TestWidgetsFlutterBinding.ensureInitialized();

  group('resolveActiveTaxpayer (MODULE-taxpayers §2.9)', () {
    test('the stored id wins while it exists and is not archived', () {
      expect(
        resolveActiveTaxpayer([_first, _second], storedId: secondTaxpayerId),
        _second,
      );
    });

    test('a stored id that is archived or gone falls back to the first '
        'non-archived by created_at', () {
      final list = [_archived, _second, _first];
      expect(resolveActiveTaxpayer(list, storedId: archivedTaxpayerId), _first);
      expect(resolveActiveTaxpayer(list, storedId: 'deleted'), _first);
      expect(resolveActiveTaxpayer(list, storedId: null), _first);
    });

    test('none when every taxpayer is archived, or there are none', () {
      expect(resolveActiveTaxpayer([_archived], storedId: null), isNull);
      expect(resolveActiveTaxpayer(const [], storedId: null), isNull);
    });
  });

  group('activeTaxpayerProvider', () {
    late FakeTaxpayersRepository repository;
    late InMemoryActiveTaxpayerStore store;
    late Session session;
    late ProviderContainer container;

    setUp(() {
      repository = FakeTaxpayersRepository()
        ..taxpayers.addAll([_first, _second, _archived]);
      store = InMemoryActiveTaxpayerStore();
      session = Session(
        store: InMemoryRefreshTokenStore(),
        rotate: (_) async => tokenPair(access: 'a2', refresh: 'r2'),
      );
      container = ProviderContainer(
        overrides: [
          sessionProvider.overrideWithValue(session),
          taxpayersRepositoryProvider.overrideWithValue(repository),
          activeTaxpayerStoreProvider.overrideWithValue(store),
        ],
      );
    });
    tearDown(() {
      container.dispose();
      session.dispose();
    });

    Future<String?> activeId() async =>
        (await container.read(activeTaxpayerProvider.future))?.id;

    test('restores the stored taxpayer after a restart', () async {
      store.id = secondTaxpayerId;

      expect(await activeId(), secondTaxpayerId);
    });

    test('keeps only the id on the device, and keeps the fallback', () async {
      store.id = 'deleted-on-another-device';

      expect(await activeId(), _first.id);
      expect(store.id, _first.id);
    });

    test('selecting a taxpayer persists it', () async {
      await activeId();

      await container
          .read(activeTaxpayerProvider.notifier)
          .select(secondTaxpayerId);

      expect(store.id, secondTaxpayerId);
      expect(await activeId(), secondTaxpayerId);
    });

    test('archiving the active one moves to the next non-archived', () async {
      store.id = _first.id;
      await activeId();

      await repository.archive(_first.id);
      await container.read(taxpayersProvider.notifier).reload();

      expect(await activeId(), secondTaxpayerId);
    });

    test('a sign-out forgets the stored id and the list', () async {
      await session.start(tokenPair(access: 'a1', refresh: 'r1'));
      store.id = secondTaxpayerId;
      await activeId();

      await session.end();

      expect(store.id, isNull);
      repository.taxpayers.clear();
      expect(await container.read(taxpayersProvider.future), isEmpty);
    });
  });

  // Gate M2-13: Riverpod reuses the notifier across rebuilds and its `ref`
  // is the current build's, so `ref.mounted` alone cannot tell a stale
  // answer. A refresh started for one account must never land on the next.
  group('stale list answers', () {
    late _HeldListRepository repository;
    late Session session;
    late ProviderContainer container;

    setUp(() {
      repository = _HeldListRepository()..taxpayers.addAll([_first]);
      session = Session(
        store: InMemoryRefreshTokenStore(),
        rotate: (_) async => tokenPair(access: 'a2', refresh: 'r2'),
      );
      container = ProviderContainer(
        overrides: [
          sessionProvider.overrideWithValue(session),
          taxpayersRepositoryProvider.overrideWithValue(repository),
          activeTaxpayerStoreProvider.overrideWithValue(
            InMemoryActiveTaxpayerStore(),
          ),
        ],
      );
    });
    tearDown(() {
      container.dispose();
      session.dispose();
    });

    test(
      'a resume refresh answered after a change of account is dropped',
      () async {
        await session.start(tokenPair(access: 'a1', refresh: 'r1'));
        await container.read(taxpayersProvider.future);
        final stale = Completer<List<Taxpayer>>();
        repository.held = stale;
        binding
          ..handleAppLifecycleStateChanged(AppLifecycleState.inactive)
          ..handleAppLifecycleStateChanged(AppLifecycleState.resumed);

        await session.end();
        repository.taxpayers
          ..clear()
          ..add(_second);
        await session.start(tokenPair(access: 'b1', refresh: 'b2'));
        expect(await container.read(taxpayersProvider.future), [_second]);
        stale.complete([_first]);
        await pumpEventQueue();

        expect(container.read(taxpayersProvider).value, [_second]);
      },
    );

    test(
      'a reload that fails keeps the last list and does not throw',
      () async {
        await container.read(taxpayersProvider.future);
        repository.failNext['list'] = const ApiUnreachable();

        await container.read(taxpayersProvider.notifier).reload();

        expect(container.read(taxpayersProvider).value, [_first]);
      },
    );
  });

  group('masked RFC', () {
    test('the server mask reads with bullets, never asterisks', () {
      expect(displayRfc('GOMA******AB6'), 'GOMA••••••AB6');
      expect(displayRfc('ABC******XY9'), 'ABC••••••XY9');
    });
  });
}
