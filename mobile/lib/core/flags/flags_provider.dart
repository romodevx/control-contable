import 'dart:developer' as developer;

import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flags_repository.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';

const _logName = 'flags';

/// The one source of module flags in the app (FlagsProvider), fed by
/// `GET /v1/flags`.
///
/// - **Fail-closed:** while the first answer is pending, or when it fails,
///   no module is enabled; keys this build does not know are dropped.
/// - **Cached:** widgets read the in-memory answer, never the network.
/// - **Refreshed on app resume.** A failed refresh keeps the last answer the
///   server gave: hiding every module because the phone woke up offline would
///   be wrong, and the server still answers 404 for a disabled module.
///
/// This is UX only; the backend 404 is the security boundary. Widgets read
/// flags through [FlagsWidgetRef.useFlag] or `FlagGate`, never this directly.
class FlagsNotifier extends AsyncNotifier<Set<AppModule>> {
  @override
  Future<Set<AppModule>> build() {
    final lifecycle = AppLifecycleListener(onResume: refresh);
    ref.onDispose(lifecycle.dispose);
    return _fetchOr(const {});
  }

  /// Asks the server again; keeps the cached answer if that fails.
  Future<void> refresh() async {
    final cached = state.value ?? const <AppModule>{};
    final fresh = await _fetchOr(cached);
    if (!ref.mounted || setEquals(fresh, state.value)) return;
    state = AsyncData(fresh);
  }

  Future<Set<AppModule>> _fetchOr(Set<AppModule> fallback) async {
    try {
      return await ref.read(flagsRepositoryProvider).fetchEnabled();
    } on Object catch (error, stackTrace) {
      developer.log(
        'GET /v1/flags failed; keeping ${fallback.length} module(s) enabled',
        name: _logName,
        error: error,
        stackTrace: stackTrace,
      );
      return fallback;
    }
  }
}

/// The enabled modules (see [FlagsNotifier]).
final AsyncNotifierProvider<FlagsNotifier, Set<AppModule>> flagsProvider =
    AsyncNotifierProvider<FlagsNotifier, Set<AppModule>>(
      FlagsNotifier.new,
      name: 'flagsProvider',
    );

/// Whether [AppModule] is enabled: `false` while loading or on error.
final ProviderFamily<bool, AppModule> moduleEnabledProvider =
    Provider.family<bool, AppModule>(
      (ref, module) => ref.watch(
        flagsProvider.select((flags) => flags.value?.contains(module) ?? false),
      ),
      name: 'moduleEnabledProvider',
    );

/// The single flag check for widgets: `ref.useFlag(AppModule.incomes)`.
extension FlagsWidgetRef on WidgetRef {
  /// Whether [module] is enabled; rebuilds the widget when that changes.
  bool useFlag(AppModule module) => watch(moduleEnabledProvider(module));
}
