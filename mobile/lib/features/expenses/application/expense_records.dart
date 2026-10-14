import 'package:control_contable/api/generated/models/expense.dart';
import 'package:control_contable/api/generated/models/expense_create.dart';
import 'package:control_contable/api/generated/models/payment_method.dart';
import 'package:control_contable/core/access/read_only_access.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/records/record_list.dart';
import 'package:control_contable/core/records/stacked_undo.dart';
import 'package:control_contable/features/expenses/application/expense_list.dart';
import 'package:control_contable/features/expenses/data/expenses_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The writes of expenses: create, edit, delete and its Undo (restore).
/// Delete and restore are allowed while read-only; restore never consumes
/// a trial slot (contract 0.6.1). After each one every expense list
/// reloads, so the totals and the other regime filters follow the server.
class ExpenseRecords {
  /// The writes over [_ref].
  new(this._ref);

  final Ref _ref;

  ExpensesRepository get _repository => _ref.read(expensesRepositoryProvider);

  /// Records [expense] for [taxpayerId]; every list reloads and the next
  /// form starts from its payment method. A `402` without a trial cap
  /// marks the session read-only. Throws the `ApiFailure`.
  Future<Expense> create(String taxpayerId, ExpenseCreate expense) async {
    try {
      final created = await _repository.create(
        taxpayerId: taxpayerId,
        expense: expense,
      );
      _ref
        ..invalidate(expenseListProvider)
        ..read(lastPaymentMethodProvider.notifier).rememberMethodOf(expense)
        ..read(categoryUsageProvider.notifier).used(expense.categoryCode);
      return created;
    } on ApiEntitlementRequired catch (failure) {
      if (failure.trialCapOf == null) {
        _ref.read(readOnlyAccessProvider.notifier).markReadOnly();
      }
      rethrow;
    }
  }

  /// Edits [expenseId] of [taxpayerId] with the merge-patch [patch]; every
  /// list reloads. A `402` (edits never hit the trial cap) marks the
  /// session read-only. Throws the `ApiFailure`.
  Future<Expense> update(
    String taxpayerId,
    String expenseId,
    Map<String, Object?> patch,
  ) async {
    try {
      final updated = await _repository.update(
        taxpayerId: taxpayerId,
        expenseId: expenseId,
        patch: patch,
      );
      _ref.invalidate(expenseListProvider);
      return updated;
    } on ApiEntitlementRequired {
      _ref.read(readOnlyAccessProvider.notifier).markReadOnly();
      rethrow;
    }
  }

  /// Deletes [expense], hiding it from [query]'s list at once. Throws the
  /// `ApiFailure` when the server refuses (the list brings it back).
  Future<void> delete(RecordListQuery query, Expense expense) async {
    _ref.read(expenseListProvider(query).notifier).hide(expense);
    try {
      await _repository.delete(
        taxpayerId: expense.taxpayerId,
        expenseId: expense.id,
      );
    } finally {
      _ref.invalidate(expenseListProvider);
    }
  }

  /// Restores every one of [expenses], going on past a refusal; no answer
  /// here marks the session read-only.
  Future<RestoreReport<ApiFailure>> restore(List<Expense> expenses) async {
    try {
      return await restoreEach<Expense, ApiFailure>(
        expenses,
        (expense) => _repository.restore(
          taxpayerId: expense.taxpayerId,
          expenseId: expense.id,
        ),
      );
    } finally {
      _ref.invalidate(expenseListProvider);
    }
  }
}

/// See [ExpenseRecords].
final expenseRecordsProvider = Provider<ExpenseRecords>(
  ExpenseRecords.new,
  name: 'expenseRecordsProvider',
);

/// How the last expense recorded in this session was paid: the next form
/// starts there (OQ-E6: debit card the first time). Memory only.
class LastPaymentMethodNotifier extends Notifier<PaymentMethod> {
  @override
  PaymentMethod build() => PaymentMethod.cardDebit;

  /// [expense] was just recorded: its payment method starts the next form.
  void rememberMethodOf(ExpenseCreate expense) => state = expense.paymentMethod;
}

/// See [LastPaymentMethodNotifier].
final NotifierProvider<LastPaymentMethodNotifier, PaymentMethod>
lastPaymentMethodProvider =
    NotifierProvider<LastPaymentMethodNotifier, PaymentMethod>(
      LastPaymentMethodNotifier.new,
      name: 'lastPaymentMethodProvider',
    );

/// How many expenses of each category this session recorded: the picker
/// lists the most used first (MODULE-expenses §7). Memory only.
class CategoryUsageNotifier extends Notifier<Map<String, int>> {
  @override
  Map<String, int> build() => const {};

  /// An expense of [categoryCode] was just recorded.
  void used(String categoryCode) =>
      state = {...state, categoryCode: (state[categoryCode] ?? 0) + 1};
}

/// See [CategoryUsageNotifier].
final NotifierProvider<CategoryUsageNotifier, Map<String, int>>
categoryUsageProvider =
    NotifierProvider<CategoryUsageNotifier, Map<String, int>>(
      CategoryUsageNotifier.new,
      name: 'categoryUsageProvider',
    );

/// The expenses the Undo toast would restore (it outlives the list screen).
final expensesUndoProvider = Provider<StackedUndo<Expense>>(
  (ref) => StackedUndo<Expense>(),
  name: 'expensesUndoProvider',
);
