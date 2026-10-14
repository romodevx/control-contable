import 'package:control_contable/core/session/session.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Whether the server has already told this session that the user is
/// read-only (`402 entitlement_required`: trial over, no plan).
///
/// The frozen contract exposes no access state before a write (`GET /v1/me`
/// carries none; `GET /v1/subscription` is M7, not frozen), so the app can
/// only learn it from a refused write. Once learned, flows that end in a
/// write (the taxpayer onboarding) show the paywall up front instead of
/// letting the user fill every step first.
///
/// UX only: the server's `402` stays the boundary. `true` is forgotten on
/// any session change (another account starts unknown) and when the app
/// resumes (a plan may have been bought in the store meanwhile).
class ReadOnlyAccessNotifier extends Notifier<bool> {
  @override
  bool build() {
    final session = ref.watch(sessionProvider)..addListener(forget);
    final lifecycle = AppLifecycleListener(onResume: forget);
    ref.onDispose(() {
      session.removeListener(forget);
      lifecycle.dispose();
    });
    return false;
  }

  /// A write was refused with `402 entitlement_required`.
  void markReadOnly() => state = true;

  /// Back to unknown: the next write asks the server again.
  void forget() => state = false;
}

/// See [ReadOnlyAccessNotifier].
final NotifierProvider<ReadOnlyAccessNotifier, bool> readOnlyAccessProvider =
    NotifierProvider<ReadOnlyAccessNotifier, bool>(
      ReadOnlyAccessNotifier.new,
      name: 'readOnlyAccessProvider',
    );
