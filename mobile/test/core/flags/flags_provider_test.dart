import 'dart:async';

import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flags_provider.dart';
import 'package:control_contable/core/flags/flags_repository.dart';
import 'package:control_contable/core/network/dio_provider.dart';
import 'package:dio/dio.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/recording_adapter.dart';

const _flagsPath = '/v1/flags';

CannedReply _enabled(List<String> keys) => (
  status: 200,
  body: {'enabled': keys},
);

ProviderContainer _containerWith(RecordingAdapter adapter) =>
    ProviderContainer.test(
      overrides: [
        dioProvider.overrideWith(
          (ref) =>
              Dio(BaseOptions(baseUrl: 'http://api.test'))
                ..httpClientAdapter = adapter,
        ),
      ],
    );

Future<void> _resumeApp(WidgetTester tester) async {
  tester.binding
    ..handleAppLifecycleStateChanged(AppLifecycleState.inactive)
    ..handleAppLifecycleStateChanged(AppLifecycleState.resumed);
  await tester.pump();
}

/// Waits for the in-flight flags request (and the state update after it).
Future<void> _settle(ProviderContainer container) async {
  await container.read(flagsProvider.future);
  await Future<void>.delayed(Duration.zero);
}

class _PendingRepository implements FlagsRepository {
  final completer = Completer<Set<AppModule>>();

  @override
  Future<Set<AppModule>> fetchEnabled() => completer.future;
}

void main() {
  testWidgets('modules the server lists are ON, every other module is OFF', (
    tester,
  ) async {
    final container = _containerWith(
      RecordingAdapter(
        replies: {
          _flagsPath: _enabled(['module_auth', 'module_incomes']),
        },
      ),
    );

    await _settle(container);

    expect(container.read(moduleEnabledProvider(AppModule.auth)), isTrue);
    expect(container.read(moduleEnabledProvider(AppModule.incomes)), isTrue);
    expect(container.read(moduleEnabledProvider(AppModule.expenses)), isFalse);
  });

  testWidgets('keys the app does not know are ignored, not a crash', (
    tester,
  ) async {
    final container = _containerWith(
      RecordingAdapter(
        replies: {
          _flagsPath: _enabled(['module_from_the_future', 'module_dashboard']),
        },
      ),
    );

    await _settle(container);

    expect(container.read(moduleEnabledProvider(AppModule.dashboard)), isTrue);
    expect(container.read(flagsProvider).value, {AppModule.dashboard});
  });

  testWidgets('fail-closed: a server error turns every module OFF', (
    tester,
  ) async {
    final container = _containerWith(
      RecordingAdapter(
        replies: {_flagsPath: (status: 500, body: <String, Object?>{})},
      ),
    );

    await _settle(container);

    for (final module in AppModule.values) {
      expect(container.read(moduleEnabledProvider(module)), isFalse);
    }
  });

  testWidgets('fail-closed: while the first answer is pending, all are OFF', (
    tester,
  ) async {
    final repository = _PendingRepository();
    final container = ProviderContainer.test(
      overrides: [flagsRepositoryProvider.overrideWithValue(repository)],
    );

    expect(container.read(moduleEnabledProvider(AppModule.auth)), isFalse);

    repository.completer.complete({AppModule.auth});
    await _settle(container);
    expect(container.read(moduleEnabledProvider(AppModule.auth)), isTrue);
  });

  testWidgets('the answer is cached and refreshed when the app resumes', (
    tester,
  ) async {
    final adapter = RecordingAdapter(
      replies: {
        _flagsPath: _enabled(['module_incomes']),
      },
    );
    final container = _containerWith(adapter);
    container.listen(moduleEnabledProvider(AppModule.incomes), (_, _) {});
    await _settle(container);

    expect(container.read(moduleEnabledProvider(AppModule.incomes)), isTrue);
    container.read(moduleEnabledProvider(AppModule.incomes));
    expect(adapter.requests, hasLength(1), reason: 'reads hit the cache');

    adapter.reply(_flagsPath, _enabled([]));
    await _resumeApp(tester);
    await _settle(container);

    expect(adapter.requests, hasLength(2));
    expect(container.read(moduleEnabledProvider(AppModule.incomes)), isFalse);
  });

  testWidgets('a failed refresh keeps the last answer from the server', (
    tester,
  ) async {
    final adapter = RecordingAdapter(
      replies: {
        _flagsPath: _enabled(['module_incomes']),
      },
    );
    final container = _containerWith(adapter);
    container.listen(moduleEnabledProvider(AppModule.incomes), (_, _) {});
    await _settle(container);

    adapter.reply(
      _flagsPath,
      DioException.connectionError(
        requestOptions: RequestOptions(path: _flagsPath),
        reason: 'offline',
      ),
    );
    await _resumeApp(tester);
    await _settle(container);

    expect(adapter.requests, hasLength(2));
    expect(container.read(moduleEnabledProvider(AppModule.incomes)), isTrue);
  });
}
