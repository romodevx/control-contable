import 'package:control_contable/api/generated/clients/incomes_client.dart';
import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/api/generated/models/income_create.dart';
import 'package:control_contable/api/generated/models/income_defaults.dart';
import 'package:control_contable/api/generated/models/income_page.dart';
import 'package:control_contable/api/generated/models/platform_activity.dart';
import 'package:control_contable/api/generated/models/withholding_preview.dart';
import 'package:control_contable/core/network/api_client_provider.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/network/dio_provider.dart';
import 'package:control_contable/features/incomes/domain/income_preview_query.dart';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The income operations of the active taxpayer (`/v1/taxpayers/{id}/incomes`).
/// Every method throws an `ApiFailure` when the server refuses or cannot be
/// reached.
abstract interface class IncomesRepository {
  /// One page of [period]'s incomes (`YYYY-MM`), narrowed to [regime] when
  /// given, with the totals of the whole filtered month.
  Future<IncomePage> list({
    required String taxpayerId,
    required String period,
    String? regime,
    String? cursor,
  });

  /// Records an income. Retrying with the same client `id` answers the
  /// stored income (`200`) instead of a second record.
  Future<Income> create({
    required String taxpayerId,
    required IncomeCreate income,
  });

  /// The IVA rates for a new income and the withholding facts of the last
  /// income in [regime] (`last_used`, MODULE-withholdings §2.1).
  Future<IncomeDefaults> defaults({
    required String taxpayerId,
    required String regime,
    String? on,
    PlatformActivity? platformActivity,
  });

  /// The withholding the server derives from [query], storing nothing
  /// (ADR-0008: the app never computes one).
  Future<WithholdingPreview> previewWithholdings({
    required String taxpayerId,
    required IncomePreviewQuery query,
  });

  /// One live income (`404` once deleted or purged).
  Future<Income> get({required String taxpayerId, required String incomeId});

  /// Edits an income with a merge-patch [patch] (top-level `updateIncome`
  /// keys): an absent key is unchanged, `null` clears a nullable field.
  Future<Income> update({
    required String taxpayerId,
    required String incomeId,
    required Map<String, Object?> patch,
  });

  /// Soft-deletes an income (allowed while read-only).
  Future<void> delete({required String taxpayerId, required String incomeId});

  /// Undoes a delete (never consumes a trial slot; allowed while read-only).
  Future<Income> restore({
    required String taxpayerId,
    required String incomeId,
  });
}

/// [IncomesRepository] over the generated client, except [update]: the
/// generated `IncomeUpdate` cannot tell "unset" from "clear" (Dart `T?`), so
/// the merge-patch travels as a hand-written map on the same [Dio]
/// (ADR-0003 "Null on the wire").
class ApiIncomesRepository implements IncomesRepository {
  /// The repository on [_client] and its transport [_dio].
  new(this._client, this._dio);

  final IncomesClient _client;
  final Dio _dio;

  @override
  Future<IncomePage> list({
    required String taxpayerId,
    required String period,
    String? regime,
    String? cursor,
  }) => guardApi(
    () => _client.listIncomes(
      taxpayerId: taxpayerId,
      period: period,
      regime: regime,
      cursor: cursor,
    ),
  );

  @override
  Future<Income> create({
    required String taxpayerId,
    required IncomeCreate income,
  }) => guardApi(
    () => _client.createIncome(taxpayerId: taxpayerId, body: income),
  );

  @override
  Future<IncomeDefaults> defaults({
    required String taxpayerId,
    required String regime,
    String? on,
    PlatformActivity? platformActivity,
  }) => guardApi(
    () => _client.getIncomeDefaults(
      taxpayerId: taxpayerId,
      regime: regime,
      onValue: on,
      platformActivity: platformActivity,
    ),
  );

  @override
  Future<WithholdingPreview> previewWithholdings({
    required String taxpayerId,
    required IncomePreviewQuery query,
  }) => guardApi(
    () => _client.previewIncomeWithholdings(
      taxpayerId: taxpayerId,
      regime: query.regime,
      onValue: query.on,
      amountCents: query.amountCents,
      ivaRate: query.ivaRate,
      ivaCents: query.ivaCents,
      counterpartyKind: query.counterpartyKind,
      operationKind: query.operationKind,
      residentialRent: query.residentialRent,
      viaPlatform: query.viaPlatform,
      platformActivity: query.platformActivity,
    ),
  );

  @override
  Future<Income> get({required String taxpayerId, required String incomeId}) =>
      guardApi(
        () => _client.getIncome(taxpayerId: taxpayerId, incomeId: incomeId),
      );

  @override
  Future<Income> update({
    required String taxpayerId,
    required String incomeId,
    required Map<String, Object?> patch,
  }) => guardApi(() async {
    final response = await _dio.patch<Map<String, Object?>>(
      '/v1/taxpayers/${Uri.encodeComponent(taxpayerId)}'
      '/incomes/${Uri.encodeComponent(incomeId)}',
      data: patch,
    );
    return Income.fromJson(response.data!);
  });

  @override
  Future<void> delete({required String taxpayerId, required String incomeId}) =>
      guardApi(
        () => _client.deleteIncome(taxpayerId: taxpayerId, incomeId: incomeId),
      );

  @override
  Future<Income> restore({
    required String taxpayerId,
    required String incomeId,
  }) => guardApi(
    () => _client.restoreIncome(taxpayerId: taxpayerId, incomeId: incomeId),
  );
}

/// The income operations. Tests override it with a fake.
final incomesRepositoryProvider = Provider<IncomesRepository>(
  (ref) => ApiIncomesRepository(
    ref.watch(apiClientProvider).incomes,
    ref.watch(dioProvider),
  ),
  name: 'incomesRepositoryProvider',
);
