import 'dart:async';

import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/taxpayers/data/active_taxpayer_store.dart';
import 'package:control_contable/features/taxpayers/data/taxpayers_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Every taxpayer of the signed-in user, archived ones too (at most 10),
/// in `created_at` order: the one source the switcher, the settings and
/// the active taxpayer read. Each write in settings reloads it.
///
/// A session change (sign-in, sign-out, another account) reloads it; a
/// sign-out also forgets the device's active-taxpayer id, so the next
/// account never starts on someone else's choice (RLS answers `404` for it
/// anyway). A failed load shows a retry: no automatic retry.
class TaxpayersNotifier extends AsyncNotifier<List<Taxpayer>> {
  @override
  Future<List<Taxpayer>> build() {
    final session = ref.watch(sessionProvider);
    var status = session.status;
    void onSessionChanged() {
      if (session.status == status) return;
      status = session.status;
      if (status == SessionStatus.signedOut) {
        unawaited(ref.read(activeTaxpayerStoreProvider).clear());
      }
      ref.invalidateSelf();
    }

    session.addListener(onSessionChanged);
    ref.onDispose(() => session.removeListener(onSessionChanged));
    return ref.read(taxpayersRepositoryProvider).listTaxpayers();
  }

  /// Loads the list again, keeping the current one on screen meanwhile.
  Future<void> reload() async {
    ref.invalidateSelf();
    await future;
  }
}

/// See [TaxpayersNotifier].
final AsyncNotifierProvider<TaxpayersNotifier, List<Taxpayer>>
taxpayersProvider = AsyncNotifierProvider<TaxpayersNotifier, List<Taxpayer>>(
  TaxpayersNotifier.new,
  name: 'taxpayersProvider',
  retry: _noAutomaticRetry,
);

Duration? _noAutomaticRetry(int retryCount, Object error) => null;
