import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/core/access/read_only_access.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/records/stacked_undo.dart';
import 'package:control_contable/features/incomes/application/income_list.dart';
import 'package:control_contable/features/incomes/data/incomes_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The writes of the income list: delete (allowed while read-only) and its
/// Undo (restore: never consumes a trial slot, `402` while read-only).
/// After each one every income list reloads, so the totals and the other
/// regime filters follow the server.
class IncomeRecords {
  /// The writes over [_ref].
  new(this._ref);

  final Ref _ref;

  IncomesRepository get _repository => _ref.read(incomesRepositoryProvider);

  /// Whether the server already said this account is read-only: a delete
  /// is then one-way (its Undo would be refused).
  bool get knownReadOnly => _ref.read(readOnlyAccessProvider);

  /// Deletes [income], hiding it from [query]'s list at once. Throws the
  /// `ApiFailure` when the server refuses (the list brings it back).
  Future<void> delete(IncomeListQuery query, Income income) async {
    _ref.read(incomeListProvider(query).notifier).hide(income);
    try {
      await _repository.delete(
        taxpayerId: income.taxpayerId,
        incomeId: income.id,
      );
    } finally {
      _ref.invalidate(incomeListProvider);
    }
  }

  /// Restores [incomes] in order; stops at the first refusal and throws
  /// it. A `402` marks the session read-only.
  Future<void> restore(List<Income> incomes) async {
    try {
      for (final income in incomes) {
        await _repository.restore(
          taxpayerId: income.taxpayerId,
          incomeId: income.id,
        );
      }
    } on ApiEntitlementRequired {
      _ref.read(readOnlyAccessProvider.notifier).markReadOnly();
      rethrow;
    } finally {
      _ref.invalidate(incomeListProvider);
    }
  }
}

/// See [IncomeRecords].
final incomeRecordsProvider = Provider<IncomeRecords>(
  IncomeRecords.new,
  name: 'incomeRecordsProvider',
);

/// The incomes the Undo toast would restore (it outlives the list screen).
final incomesUndoProvider = Provider<StackedUndo<Income>>(
  (ref) => StackedUndo<Income>(),
  name: 'incomesUndoProvider',
);
