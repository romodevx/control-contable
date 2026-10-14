import 'package:control_contable/core/diagnostics/error_reporter.dart';

/// An [ErrorReporter] that keeps every report, so tests can check what an
/// absorbed failure left behind.
class RecordingErrorReporter implements ErrorReporter {
  /// Every report, in order.
  final List<({String operation, Object error})> reports = [];

  @override
  void report(String operation, Object error) =>
      reports.add((operation: operation, error: error));
}
