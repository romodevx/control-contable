// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

import '../models/expense.dart';
import '../models/expense_category_list.dart';
import '../models/expense_create.dart';
import '../models/expense_page.dart';
import '../models/expense_update.dart';
import '../models/period.dart';
import '../models/regime_code.dart';
import '../models/uuid.dart';

part 'expenses_client.g.dart';

@RestApi()
abstract class ExpensesClient {
  factory ExpensesClient(Dio dio, {String? baseUrl}) = _ExpensesClient;

  /// Expense categories with deductibility hints per regime
  @GET('/v1/expense-categories')
  Future<ExpenseCategoryList> listExpenseCategories({
    @Query('regime') RegimeCode? regime,
  });

  /// List a month's expenses (keyset pagination) with period totals.
  ///
  /// [cursor] - Opaque keyset cursor from `next_cursor`.
  @GET('/v1/taxpayers/{taxpayerId}/expenses')
  Future<ExpensePage> listExpenses({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Query('period') required Period period,
    @Query('limit') int? limit = 50,
    @Query('regime') RegimeCode? regime,
    @Query('cursor') String? cursor,
  });

  /// Record an expense (trial cap enforced)
  @POST('/v1/taxpayers/{taxpayerId}/expenses')
  Future<Expense> createExpense({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Body() required ExpenseCreate body,
  });

  /// Get an expense
  @GET('/v1/taxpayers/{taxpayerId}/expenses/{expenseId}')
  Future<Expense> getExpense({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('expenseId') required Uuid expenseId,
  });

  /// Edit an expense
  @PATCH('/v1/taxpayers/{taxpayerId}/expenses/{expenseId}')
  Future<Expense> updateExpense({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('expenseId') required Uuid expenseId,
    @Body() required ExpenseUpdate body,
  });

  /// Soft-delete an expense (undo via restore)
  @DELETE('/v1/taxpayers/{taxpayerId}/expenses/{expenseId}')
  Future<void> deleteExpense({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('expenseId') required Uuid expenseId,
  });

  /// Undo a delete
  @POST('/v1/taxpayers/{taxpayerId}/expenses/{expenseId}/restore')
  Future<Expense> restoreExpense({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('expenseId') required Uuid expenseId,
  });
}
