import 'package:control_contable/api/generated/models/history_year.dart';
import 'package:control_contable/features/history/data/history_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';

/// Which year of History to read: a taxpayer and a fiscal year.
typedef HistoryQuery = ({String taxpayerId, int fiscalYear});

/// One fiscal year of one taxpayer, keyed by both: switching taxpayer or
/// year reads a new key, so another taxpayer's figures are never shown
/// while it loads. Derived on every read (ADR-0002 §5); a failure shows a
/// retry, never an automatic one.
final FutureProviderFamily<HistoryYear, HistoryQuery> historyYearProvider =
    FutureProvider.autoDispose.family<HistoryYear, HistoryQuery>(
      (ref, query) => ref
          .read(historyRepositoryProvider)
          .getHistory(
            taxpayerId: query.taxpayerId,
            fiscalYear: query.fiscalYear,
          ),
      name: 'historyYearProvider',
      retry: _noAutomaticRetry,
    );

Duration? _noAutomaticRetry(int retryCount, Object error) => null;

/// The fiscal year chosen on History: `null` is the current one (the
/// default). Forgotten when the screen closes.
class HistoryYearNotifier extends Notifier<int?> {
  @override
  int? build() => null;

  /// Chooses [year]; choosing [current] goes back to `null`.
  void select(int year, {required int current}) =>
      state = year == current ? null : year;
}

/// See [HistoryYearNotifier].
final NotifierProvider<HistoryYearNotifier, int?> historyChosenYearProvider =
    NotifierProvider.autoDispose<HistoryYearNotifier, int?>(
      HistoryYearNotifier.new,
      name: 'historyChosenYearProvider',
    );
