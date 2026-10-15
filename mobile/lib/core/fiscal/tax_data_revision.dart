import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Counts the writes to the data every tax figure is derived from (the
/// income and expense records, the fiscal inputs) in this app run.
///
/// It is the single invalidation path between the modules: a write bumps
/// it, and each derived read (History, Home's dashboard, the estimate and
/// its breakdown) watches it, so every screen on the stack reads again,
/// whichever screen, route or app-wide toast (*Deshacer*) the write came
/// from. The value itself means nothing; only its change does.
class TaxDataRevision extends Notifier<int> {
  @override
  int build() => 0;

  /// Marks the derived figures stale after a write (or a failed one: the
  /// server's state is then unknown, so reading again is the safe side).
  void bump() => state++;
}

/// See [TaxDataRevision].
final NotifierProvider<TaxDataRevision, int> taxDataRevisionProvider =
    NotifierProvider<TaxDataRevision, int>(
      TaxDataRevision.new,
      name: 'taxDataRevisionProvider',
    );
