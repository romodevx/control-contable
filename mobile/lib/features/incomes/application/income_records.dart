import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/api/generated/models/income_create.dart';
import 'package:control_contable/core/access/read_only_access.dart';
import 'package:control_contable/core/fiscal/tax_data_revision.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/records/stacked_undo.dart';
import 'package:control_contable/core/session/session_scoped.dart';
import 'package:control_contable/features/incomes/application/income_list.dart';
import 'package:control_contable/features/incomes/data/incomes_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The writes of incomes: create, edit, delete and its Undo (restore).
/// Delete and restore are allowed while read-only; restore never consumes
/// a trial slot.
/// After each one every income list reloads, so the totals and the other
/// regime filters follow the server, and the tax-data revision moves, so
/// every figure derived from the records (History, Home, the estimate)
/// reads again wherever it is on the stack.
class IncomeRecords {
  /// The writes over [_ref].
  new(this._ref);

  final Ref _ref;

  IncomesRepository get _repository => _ref.read(incomesRepositoryProvider);

  void _changed() {
    _ref.invalidate(incomeListProvider);
    _ref.read(taxDataRevisionProvider.notifier).bump();
  }

  /// Records [income] for [taxpayerId]; every list reloads. A `402`
  /// without a trial cap marks the session read-only (a spent trial cap
  /// does not: other writes still work). Throws the `ApiFailure`.
  Future<Income> create(String taxpayerId, IncomeCreate income) async {
    try {
      final created = await _repository.create(
        taxpayerId: taxpayerId,
        income: income,
      );
      _changed();
      return created;
    } on ApiEntitlementRequired catch (failure) {
      if (failure.trialCapOf == null) {
        _ref.read(readOnlyAccessProvider.notifier).markReadOnly();
      }
      rethrow;
    }
  }

  /// Edits [incomeId] of [taxpayerId] with the merge-patch [patch]; every
  /// list reloads. A `402` (edits never hit the trial cap) marks the
  /// session read-only. Throws the `ApiFailure`.
  Future<Income> update(
    String taxpayerId,
    String incomeId,
    Map<String, Object?> patch,
  ) async {
    try {
      final updated = await _repository.update(
        taxpayerId: taxpayerId,
        incomeId: incomeId,
        patch: patch,
      );
      _changed();
      return updated;
    } on ApiEntitlementRequired {
      _ref.read(readOnlyAccessProvider.notifier).markReadOnly();
      rethrow;
    }
  }

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
      _changed();
    }
  }

  /// Restores every one of [incomes], going on past a refusal. Allowed
  /// while read-only (contract 0.6.1), so no answer here marks the session.
  Future<RestoreReport<ApiFailure>> restore(List<Income> incomes) async {
    try {
      return await restoreEach<Income, ApiFailure>(
        incomes,
        (income) => _repository.restore(
          taxpayerId: income.taxpayerId,
          incomeId: income.id,
        ),
      );
    } finally {
      _changed();
    }
  }
}

/// See [IncomeRecords].
final incomeRecordsProvider = Provider<IncomeRecords>(
  IncomeRecords.new,
  name: 'incomeRecordsProvider',
);

/// The incomes the Undo toast would restore (it outlives the list screen).
final incomesUndoProvider = Provider<StackedUndo<Income>>((ref) {
  forgetOnSignOut(ref);
  return StackedUndo<Income>();
}, name: 'incomesUndoProvider');
