import 'package:control_contable/api/generated/clients/history_client.dart';
import 'package:control_contable/api/generated/models/history_viewed_request.dart';
import 'package:control_contable/api/generated/models/history_year.dart';
import 'package:control_contable/core/network/api_client_provider.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The History operations of M8 (`/v1/taxpayers/{id}/history/**`). Every
/// method throws an `ApiFailure` when the server refuses or cannot be
/// reached.
abstract interface class HistoryRepository {
  /// Every month of [fiscalYear] for [taxpayerId], newest first, with its
  /// totals, status and changed marker (`GET …/history`).
  Future<HistoryYear> getHistory({
    required String taxpayerId,
    required int fiscalYear,
  });

  /// Records that the user saw [period] (`YYYY-MM`) as [request] describes
  /// it (`POST …/history/{period}/viewed`). Idempotent.
  Future<void> markViewed({
    required String taxpayerId,
    required String period,
    required HistoryViewedRequest request,
  });
}

/// [HistoryRepository] over the generated client.
class ApiHistoryRepository implements HistoryRepository {
  /// The repository on [_client].
  new(this._client);

  final HistoryClient _client;

  @override
  Future<HistoryYear> getHistory({
    required String taxpayerId,
    required int fiscalYear,
  }) => guardApi(
    () => _client.getHistory(taxpayerId: taxpayerId, fiscalYear: fiscalYear),
  );

  @override
  Future<void> markViewed({
    required String taxpayerId,
    required String period,
    required HistoryViewedRequest request,
  }) => guardApi(
    () => _client.markHistoryViewed(
      taxpayerId: taxpayerId,
      period: period,
      body: request,
    ),
  );
}

/// The History operations. Tests override them with a fake.
final historyRepositoryProvider = Provider<HistoryRepository>(
  (ref) => ApiHistoryRepository(ref.watch(apiClientProvider).history),
  name: 'historyRepositoryProvider',
);
