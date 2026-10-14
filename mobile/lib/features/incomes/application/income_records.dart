import 'package:control_contable/api/generated/models/counterparty_type.dart';
import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/api/generated/models/income_create.dart';
import 'package:control_contable/core/access/read_only_access.dart';
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
/// regime filters follow the server.
class IncomeRecords {
  /// The writes over [_ref].
  new(this._ref);

  final Ref _ref;

  IncomesRepository get _repository => _ref.read(incomesRepositoryProvider);

  /// Records [income] for [taxpayerId]; every list reloads. A `402`
  /// without a trial cap marks the session read-only (a spent trial cap
  /// does not: other writes still work). Throws the `ApiFailure`.
  Future<Income> create(String taxpayerId, IncomeCreate income) async {
    try {
      final created = await _repository.create(
        taxpayerId: taxpayerId,
        income: income,
      );
      _ref
        ..invalidate(incomeListProvider)
        ..read(lastCounterpartyProvider.notifier).rememberPayerOf(income);
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
      _ref.invalidate(incomeListProvider);
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
      _ref.invalidate(incomeListProvider);
    }
  }

  /// Restores [incomes] in order; stops at the first refusal and throws
  /// it. Allowed while read-only (contract 0.6.1), so no answer here marks
  /// the session.
  Future<void> restore(List<Income> incomes) async {
    try {
      for (final income in incomes) {
        await _repository.restore(
          taxpayerId: income.taxpayerId,
          incomeId: income.id,
        );
      }
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

/// Who paid the last income recorded in this session: the next form starts
/// there (US-I1; persona física the first time). Memory only, forgotten on
/// sign-out.
class LastCounterpartyNotifier extends Notifier<CounterpartyType> {
  @override
  CounterpartyType build() {
    forgetOnSignOut(ref);
    return CounterpartyType.personaFisica;
  }

  /// [income] was just recorded: its payer starts the next form.
  void rememberPayerOf(IncomeCreate income) => state = income.counterpartyType;
}

/// See [LastCounterpartyNotifier].
final NotifierProvider<LastCounterpartyNotifier, CounterpartyType>
lastCounterpartyProvider =
    NotifierProvider<LastCounterpartyNotifier, CounterpartyType>(
      LastCounterpartyNotifier.new,
      name: 'lastCounterpartyProvider',
    );

/// The incomes the Undo toast would restore (it outlives the list screen).
final incomesUndoProvider = Provider<StackedUndo<Income>>((ref) {
  forgetOnSignOut(ref);
  return StackedUndo<Income>();
}, name: 'incomesUndoProvider');
