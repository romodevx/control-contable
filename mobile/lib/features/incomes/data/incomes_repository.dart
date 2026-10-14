import 'package:control_contable/api/generated/clients/incomes_client.dart';
import 'package:control_contable/api/generated/models/counterparty_type.dart';
import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/api/generated/models/income_create.dart';
import 'package:control_contable/api/generated/models/income_defaults.dart';
import 'package:control_contable/api/generated/models/income_page.dart';
import 'package:control_contable/api/generated/models/platform_activity.dart';
import 'package:control_contable/core/network/api_client_provider.dart';
import 'package:control_contable/core/network/api_failure.dart';
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

  /// The IVA rates and suggested withholdings for a new income.
  Future<IncomeDefaults> defaults({
    required String taxpayerId,
    required String regime,
    required CounterpartyType counterparty,
    String? on,
    PlatformActivity? platformActivity,
  });

  /// Soft-deletes an income (allowed while read-only).
  Future<void> delete({required String taxpayerId, required String incomeId});

  /// Undoes a delete (never consumes a trial slot; `402` while read-only).
  Future<Income> restore({
    required String taxpayerId,
    required String incomeId,
  });
}

/// [IncomesRepository] over the generated client.
class ApiIncomesRepository implements IncomesRepository {
  /// The repository on [_client].
  new(this._client);

  final IncomesClient _client;

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
    required CounterpartyType counterparty,
    String? on,
    PlatformActivity? platformActivity,
  }) => guardApi(
    () => _client.getIncomeDefaults(
      taxpayerId: taxpayerId,
      regime: regime,
      counterpartyType: counterparty,
      onValue: on,
      platformActivity: platformActivity,
    ),
  );

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
  (ref) => ApiIncomesRepository(ref.watch(apiClientProvider).incomes),
  name: 'incomesRepositoryProvider',
);
