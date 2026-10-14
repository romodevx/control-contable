import 'package:control_contable/api/generated/models/counterparty_type.dart';
import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/api/generated/models/income_create.dart';
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
        ..read(lastCounterpartyProvider.notifier).remembered =
            income.counterpartyType;
      return created;
    } on ApiEntitlementRequired catch (failure) {
      if (failure.trialCapOf == null) {
        _ref.read(readOnlyAccessProvider.notifier).markReadOnly();
      }
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

/// Who paid the last income recorded in this session: the next form starts
/// there (US-I1; persona física the first time). Memory only.
class LastCounterpartyNotifier extends Notifier<CounterpartyType> {
  @override
  CounterpartyType build() => CounterpartyType.personaFisica;

  /// Who paid the income just recorded.
  CounterpartyType get remembered => state;

  set remembered(CounterpartyType payer) => state = payer;
}

/// See [LastCounterpartyNotifier].
final NotifierProvider<LastCounterpartyNotifier, CounterpartyType>
lastCounterpartyProvider =
    NotifierProvider<LastCounterpartyNotifier, CounterpartyType>(
      LastCounterpartyNotifier.new,
      name: 'lastCounterpartyProvider',
    );

/// The incomes the Undo toast would restore (it outlives the list screen).
final incomesUndoProvider = Provider<StackedUndo<Income>>(
  (ref) => StackedUndo<Income>(),
  name: 'incomesUndoProvider',
);
