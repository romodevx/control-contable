import 'package:control_contable/core/flags/app_module.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Placeholder: implemented after the failing test (TDD).
abstract interface class FlagsRepository {
  Future<Set<AppModule>> fetchEnabled();
}

final flagsRepositoryProvider = Provider<FlagsRepository>(
  (ref) => throw UnimplementedError(),
);
