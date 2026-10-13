import 'dart:async';

import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flags_repository.dart';

/// A [FlagsRepository] that answers with a fixed set, or never answers.
class FakeFlagsRepository implements FlagsRepository {
  /// Answers every fetch with [enabled].
  new(Set<AppModule> enabled) : _answer = Future.value(enabled);

  /// Never answers: the flags stay loading.
  new never() : _answer = Completer<Set<AppModule>>().future;

  final Future<Set<AppModule>> _answer;

  @override
  Future<Set<AppModule>> fetchEnabled() => _answer;
}
