import 'package:control_contable/api/generated/clients/expenses_client.dart';
import 'package:control_contable/api/generated/models/expense.dart';
import 'package:control_contable/api/generated/models/expense_category_list.dart';
import 'package:control_contable/api/generated/models/expense_create.dart';
import 'package:control_contable/api/generated/models/expense_page.dart';
import 'package:control_contable/api/generated/models/withholding_preview.dart';
import 'package:control_contable/core/network/api_client_provider.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/network/dio_provider.dart';
import 'package:control_contable/features/expenses/domain/expense_preview_query.dart';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The expense operations of the active taxpayer
/// (`/v1/taxpayers/{id}/expenses`, `…/expense-categories`). Every method
/// throws an `ApiFailure` when the server refuses or cannot be reached.
abstract interface class ExpensesRepository {
  /// One page of [period]'s expenses (`YYYY-MM`, by `paid_on`), narrowed to
  /// [regime] when given, with the month's totals and the list legend.
  Future<ExpensePage> list({
    required String taxpayerId,
    required String period,
    String? regime,
    String? cursor,
  });

  /// The category catalog resolved for [regime] on [on] (`YYYY-MM-DD`,
  /// today when `null`): hints, allowed IVA rates, the cash limit and the
  /// mandatory legend.
  Future<ExpenseCategoryList> categories({
    required String taxpayerId,
    required String regime,
    String? on,
  });

  /// Records an expense. Retrying with the same client `id` answers the
  /// stored expense (`200`) instead of a second record.
  Future<Expense> create({
    required String taxpayerId,
    required ExpenseCreate expense,
  });

  /// What a persona moral must withhold from the supplier, derived by the
  /// server from [query] and stored nowhere (ADR-0008).
  Future<WithholdingPreview> previewWithholdings({
    required String taxpayerId,
    required ExpensePreviewQuery query,
  });

  /// One live expense (`404` once deleted or purged).
  Future<Expense> get({required String taxpayerId, required String expenseId});

  /// Edits an expense with a merge-patch [patch] (top-level `updateExpense`
  /// keys): an absent key is unchanged, `null` clears a nullable field or
  /// the deductibility override.
  Future<Expense> update({
    required String taxpayerId,
    required String expenseId,
    required Map<String, Object?> patch,
  });

  /// Soft-deletes an expense (allowed while read-only).
  Future<void> delete({required String taxpayerId, required String expenseId});

  /// Undoes a delete (never consumes a trial slot, never `402`).
  Future<Expense> restore({
    required String taxpayerId,
    required String expenseId,
  });
}

/// [ExpensesRepository] over the generated client, except [update]: the
/// generated `ExpenseUpdate` cannot tell "unset" from "clear" (Dart `T?`),
/// so the merge-patch travels as a hand-written map on the same [Dio]
/// (ADR-0003 "Null on the wire").
class ApiExpensesRepository implements ExpensesRepository {
  /// The repository on [_client] and its transport [_dio].
  new(this._client, this._dio);

  final ExpensesClient _client;
  final Dio _dio;

  @override
  Future<ExpensePage> list({
    required String taxpayerId,
    required String period,
    String? regime,
    String? cursor,
  }) => guardApi(
    () => _client.listExpenses(
      taxpayerId: taxpayerId,
      period: period,
      regime: regime,
      cursor: cursor,
    ),
  );

  @override
  Future<ExpenseCategoryList> categories({
    required String taxpayerId,
    required String regime,
    String? on,
  }) => guardApi(
    () => _client.listExpenseCategories(
      taxpayerId: taxpayerId,
      regime: regime,
      onValue: on,
    ),
  );

  @override
  Future<Expense> create({
    required String taxpayerId,
    required ExpenseCreate expense,
  }) => guardApi(
    () => _client.createExpense(taxpayerId: taxpayerId, body: expense),
  );

  @override
  Future<WithholdingPreview> previewWithholdings({
    required String taxpayerId,
    required ExpensePreviewQuery query,
  }) => guardApi(
    () => _client.previewExpenseWithholdings(
      taxpayerId: taxpayerId,
      regime: query.regime,
      onValue: query.on,
      amountCents: query.amountCents,
      ivaRate: query.ivaRate,
      ivaCents: query.ivaCents,
      categoryCode: query.categoryCode,
      supplierKind: query.supplierKind,
      supplierRegime: query.supplierRegime,
      operationKind: query.operationKind,
      residentialRent: query.residentialRent,
    ),
  );

  @override
  Future<Expense> get({
    required String taxpayerId,
    required String expenseId,
  }) => guardApi(
    () => _client.getExpense(taxpayerId: taxpayerId, expenseId: expenseId),
  );

  @override
  Future<Expense> update({
    required String taxpayerId,
    required String expenseId,
    required Map<String, Object?> patch,
  }) => guardApi(() async {
    final response = await _dio.patch<Map<String, Object?>>(
      '/v1/taxpayers/${Uri.encodeComponent(taxpayerId)}'
      '/expenses/${Uri.encodeComponent(expenseId)}',
      data: patch,
    );
    return Expense.fromJson(response.data!);
  });

  @override
  Future<void> delete({
    required String taxpayerId,
    required String expenseId,
  }) => guardApi(
    () => _client.deleteExpense(taxpayerId: taxpayerId, expenseId: expenseId),
  );

  @override
  Future<Expense> restore({
    required String taxpayerId,
    required String expenseId,
  }) => guardApi(
    () => _client.restoreExpense(taxpayerId: taxpayerId, expenseId: expenseId),
  );
}

/// The expense operations. Tests override it with a fake.
final expensesRepositoryProvider = Provider<ExpensesRepository>(
  (ref) => ApiExpensesRepository(
    ref.watch(apiClientProvider).expenses,
    ref.watch(dioProvider),
  ),
  name: 'expensesRepositoryProvider',
);
