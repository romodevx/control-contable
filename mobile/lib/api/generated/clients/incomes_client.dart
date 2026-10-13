// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

import '../models/income.dart';
import '../models/income_create.dart';
import '../models/income_page.dart';
import '../models/income_update.dart';
import '../models/period.dart';
import '../models/regime_code.dart';
import '../models/uuid.dart';

part 'incomes_client.g.dart';

@RestApi()
abstract class IncomesClient {
  factory IncomesClient(Dio dio, {String? baseUrl}) = _IncomesClient;

  /// List a month's incomes (keyset pagination) with period totals.
  ///
  /// [cursor] - Opaque keyset cursor from `next_cursor`.
  @GET('/v1/taxpayers/{taxpayerId}/incomes')
  Future<IncomePage> listIncomes({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Query('period') required Period period,
    @Query('limit') int? limit = 50,
    @Query('regime') RegimeCode? regime,
    @Query('cursor') String? cursor,
  });

  /// Record an income (trial cap enforced)
  @POST('/v1/taxpayers/{taxpayerId}/incomes')
  Future<Income> createIncome({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Body() required IncomeCreate body,
  });

  /// Get an income
  @GET('/v1/taxpayers/{taxpayerId}/incomes/{incomeId}')
  Future<Income> getIncome({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('incomeId') required Uuid incomeId,
  });

  /// Edit an income (402 when access is read-only)
  @PATCH('/v1/taxpayers/{taxpayerId}/incomes/{incomeId}')
  Future<Income> updateIncome({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('incomeId') required Uuid incomeId,
    @Body() required IncomeUpdate body,
  });

  /// Soft-delete an income (undo via restore)
  @DELETE('/v1/taxpayers/{taxpayerId}/incomes/{incomeId}')
  Future<void> deleteIncome({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('incomeId') required Uuid incomeId,
  });

  /// Undo a delete
  @POST('/v1/taxpayers/{taxpayerId}/incomes/{incomeId}/restore')
  Future<Income> restoreIncome({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('incomeId') required Uuid incomeId,
  });
}
