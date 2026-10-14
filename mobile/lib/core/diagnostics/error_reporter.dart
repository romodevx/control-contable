import 'dart:developer' as developer;

import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Where the app reports a failure it chose to absorb (a best-effort step
/// whose error must not reach the user), so none is dropped silently.
abstract interface class ErrorReporter {
  /// Reports that [operation] failed with [error]. Implementations keep
  /// only the error's type: its message may echo a token or personal data.
  void report(String operation, Object error);
}

/// The one line a report leaves: the [operation] and the type of [error],
/// never its message.
String describeAbsorbedError(String operation, Object error) =>
    '$operation failed: ${error.runtimeType}';

/// [ErrorReporter] over `dart:developer` logging (the debug console; a
/// release build keeps no log). A crash-reporting backend replaces it here.
class LogErrorReporter implements ErrorReporter {
  /// The reporter.
  const new();

  static const String _logName = 'control_contable.errors';

  @override
  void report(String operation, Object error) =>
      developer.log(describeAbsorbedError(operation, error), name: _logName);
}

/// The app's error reporter. Tests override it with a recording one.
final errorReporterProvider = Provider<ErrorReporter>(
  (ref) => const LogErrorReporter(),
  name: 'errorReporterProvider',
);
