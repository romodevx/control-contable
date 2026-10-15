import 'dart:io';

import 'package:flutter_test/flutter_test.dart';

/// The generated `DeviceCreate.toString()` prints `pushToken` (freezed), and
/// generated code cannot carry an override. So the token is kept in a
/// `PushToken` (redacted `toString`) everywhere, and a `DeviceCreate` exists
/// only inline at the one API call. These guards keep it that way: a new
/// construction site, or a logger that prints request bodies, fails here.
const String _sourceRoot = 'lib';
const String _generatedRoot = 'lib/api/generated';
const String _onlyAllowedSite =
    'lib/features/reminders/data/reminders_repository.dart';

Iterable<File> _handWrittenSources() =>
    Directory(_sourceRoot)
        .listSync(recursive: true)
        .whereType<File>()
        .where((file) => file.path.endsWith('.dart'))
        .where((file) => !file.path.startsWith(_generatedRoot));

List<String> _filesMatching(Pattern pattern) => [
  for (final file in _handWrittenSources())
    if (file.readAsStringSync().contains(pattern)) file.path,
];

void main() {
  test('a DeviceCreate is built only in the reminders repository', () {
    expect(_filesMatching(RegExp(r'\bDeviceCreate\(')), [_onlyAllowedSite]);
  });

  test('the raw token field is named only there too', () {
    expect(_filesMatching(RegExp(r'\bpushToken\b')), [_onlyAllowedSite]);
  });

  test('no Dio LogInterceptor prints request bodies', () {
    expect(_filesMatching('LogInterceptor'), isEmpty);
  });

  test('nothing prints with print/debugPrint in the reminders slice', () {
    final printing = _filesMatching(RegExp(r'\b(debugPrint|print)\('));
    expect(
      printing.where((path) => path.contains('features/reminders')),
      isEmpty,
    );
  });
}
