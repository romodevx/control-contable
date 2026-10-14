import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/features/taxpayers/application/taxpayer_list.dart';
import 'package:control_contable/features/taxpayers/data/active_taxpayer_store.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The active taxpayer among [taxpayers] (MODULE-taxpayers §2.9): the one
/// kept as [storedId] while it exists and is not archived; else the first
/// non-archived by `created_at`; `null` when none is left.
Taxpayer? resolveActiveTaxpayer(
  List<Taxpayer> taxpayers, {
  required String? storedId,
}) {
  final open = taxpayers.where((t) => t.archivedAt == null).toList()
    ..sort((a, b) => a.createdAt.compareTo(b.createdAt));
  return open.where((t) => t.id == storedId).firstOrNull ?? open.firstOrNull;
}

/// The taxpayer every later module scopes to (its id is the path parameter
/// of M3–M9 calls). Client state only: the id is kept on the device
/// ([ActiveTaxpayerStore]); there is no server endpoint.
///
/// Follows [taxpayersProvider]: archiving or deleting the active one moves
/// to the next non-archived; `null` sends a new user to onboarding.
class ActiveTaxpayerNotifier extends AsyncNotifier<Taxpayer?> {
  @override
  Future<Taxpayer?> build() async {
    final taxpayers = await ref.watch(taxpayersProvider.future);
    final store = ref.read(activeTaxpayerStoreProvider);
    final storedId = await store.read();
    final active = resolveActiveTaxpayer(taxpayers, storedId: storedId);
    if (active != null && active.id != storedId) await store.write(active.id);
    return active;
  }

  /// Makes [taxpayerId] the active taxpayer (also after an app restart).
  Future<void> select(String taxpayerId) async {
    await ref.read(activeTaxpayerStoreProvider).write(taxpayerId);
    ref.invalidateSelf();
    await future;
  }

  /// A taxpayer was just created: it becomes the active one once the list
  /// holds it.
  Future<void> adopt(String taxpayerId) async {
    await ref.read(activeTaxpayerStoreProvider).write(taxpayerId);
    await ref.read(taxpayersProvider.notifier).reload();
  }
}

/// See [ActiveTaxpayerNotifier].
final AsyncNotifierProvider<ActiveTaxpayerNotifier, Taxpayer?>
activeTaxpayerProvider =
    AsyncNotifierProvider<ActiveTaxpayerNotifier, Taxpayer?>(
      ActiveTaxpayerNotifier.new,
      name: 'activeTaxpayerProvider',
      retry: _noAutomaticRetry,
    );

Duration? _noAutomaticRetry(int retryCount, Object error) => null;
