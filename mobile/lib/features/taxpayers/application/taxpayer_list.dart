import 'dart:async';
import 'dart:developer' as developer;

import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/taxpayers/data/active_taxpayer_store.dart';
import 'package:control_contable/features/taxpayers/data/taxpayers_repository.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Every taxpayer of the signed-in user, archived ones too (at most 10),
/// in `created_at` order: the one source the switcher, the settings and
/// the active taxpayer read. Each write in settings reloads it.
///
/// A session change (sign-in, sign-out, another account) reloads it; a
/// sign-out also forgets the device's active-taxpayer id, so the next
/// account never starts on someone else's choice (RLS answers `404` for it
/// anyway). A failed load shows a retry: no automatic retry.
///
/// Refreshed on app resume: a taxpayer archived or deleted on another
/// device stops being active here (MODULE-taxpayers §2.9). A failed
/// refresh keeps the last list; the next resume or write asks again.
class TaxpayersNotifier extends AsyncNotifier<List<Taxpayer>> {
  @override
  Future<List<Taxpayer>> build() {
    final lifecycle = AppLifecycleListener(onResume: _refreshOnResume);
    ref.onDispose(lifecycle.dispose);
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

  Future<void> _refreshOnResume() async {
    if (state.value == null) return;
    try {
      final fresh = await ref.read(taxpayersRepositoryProvider).listTaxpayers();
      if (ref.mounted) state = AsyncData(fresh);
    } on ApiFailure catch (failure) {
      developer.log(
        'GET /v1/taxpayers on resume failed (${failure.runtimeType}); '
        'keeping the last list',
        name: 'taxpayers',
      );
    }
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
