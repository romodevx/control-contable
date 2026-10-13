import 'dart:async';

import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flags_provider.dart';
import 'package:control_contable/core/flags/flags_repository.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

/// Answers each fetch with the next scripted answer: a module set, an
/// [Exception] to throw, or a [Completer] to answer later.
class _ScriptedRepository implements FlagsRepository {
  new(this._answers);

  final List<Object> _answers;
  int calls = 0;

  @override
  Future<Set<AppModule>> fetchEnabled() async {
    final answer = _answers[calls++];
    return switch (answer) {
      final Set<AppModule> modules => modules,
      final Completer<Set<AppModule>> pending => await pending.future,
      final Exception error => throw error,
      _ => throw StateError('bad script entry: $answer'),
    };
  }
}

ProviderContainer _containerWith(FlagsRepository repository) {
  // Listening keeps the flags alive, as the mounted app does.
  return ProviderContainer.test(
    overrides: [flagsRepositoryProvider.overrideWithValue(repository)],
  )..listen(flagsProvider, (_, _) {});
}

bool _isOn(ProviderContainer container, AppModule module) =>
    container.read(moduleEnabledProvider(module));

Future<void> _resumeApp(WidgetTester tester) async {
  tester.binding
    ..handleAppLifecycleStateChanged(AppLifecycleState.inactive)
    ..handleAppLifecycleStateChanged(AppLifecycleState.resumed);
  await tester.pump();
}

void main() {
  testWidgets('modules the server lists are ON, every other module is OFF', (
    tester,
  ) async {
    final container = _containerWith(
      _ScriptedRepository([
        {AppModule.auth, AppModule.incomes},
      ]),
    );
    await tester.pump();

    expect(_isOn(container, AppModule.auth), isTrue);
    expect(_isOn(container, AppModule.incomes), isTrue);
    expect(_isOn(container, AppModule.expenses), isFalse);
  });

  testWidgets('fail-closed: an error on the first fetch turns all OFF', (
    tester,
  ) async {
    final container = _containerWith(_ScriptedRepository([Exception('500')]));
    await tester.pump();

    expect(container.read(flagsProvider).value, isEmpty);
    for (final module in AppModule.values) {
      expect(_isOn(container, module), isFalse);
    }
  });

  testWidgets('fail-closed: while the first answer is pending, all are OFF', (
    tester,
  ) async {
    final pending = Completer<Set<AppModule>>();
    final container = _containerWith(_ScriptedRepository([pending]));
    await tester.pump();

    expect(_isOn(container, AppModule.auth), isFalse);

    pending.complete({AppModule.auth});
    await tester.pump();
    expect(_isOn(container, AppModule.auth), isTrue);
  });

  testWidgets('the answer is cached and refreshed when the app resumes', (
    tester,
  ) async {
    final repository = _ScriptedRepository([
      {AppModule.incomes},
      <AppModule>{},
    ]);
    final container = _containerWith(repository);
    await tester.pump();

    expect(_isOn(container, AppModule.incomes), isTrue);
    _isOn(container, AppModule.incomes);
    expect(repository.calls, 1, reason: 'reads are served from the cache');

    await _resumeApp(tester);

    expect(repository.calls, 2);
    expect(_isOn(container, AppModule.incomes), isFalse);
  });

  testWidgets('a failed refresh keeps the last answer from the server', (
    tester,
  ) async {
    final repository = _ScriptedRepository([
      {AppModule.incomes},
      Exception('offline'),
    ]);
    final container = _containerWith(repository);
    await tester.pump();

    await _resumeApp(tester);

    expect(repository.calls, 2);
    expect(_isOn(container, AppModule.incomes), isTrue);
  });
}
