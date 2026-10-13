import 'package:control_contable/core/flags/app_module.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Placeholder: implemented after the failing test (TDD).
class FlagsNotifier extends AsyncNotifier<Set<AppModule>> {
  @override
  Future<Set<AppModule>> build() => throw UnimplementedError();
}

final flagsProvider = AsyncNotifierProvider<FlagsNotifier, Set<AppModule>>(
  FlagsNotifier.new,
);

final moduleEnabledProvider = Provider.family<bool, AppModule>(
  (ref, module) => throw UnimplementedError(),
);

extension FlagsWidgetRef on WidgetRef {
  bool useFlag(AppModule module) => throw UnimplementedError();
}
