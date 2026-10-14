import 'package:control_contable/api/generated/clients/dashboard_client.dart';
import 'package:control_contable/api/generated/models/dashboard.dart';
import 'package:control_contable/core/network/api_client_provider.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The dashboard read of M6 (`GET /v1/taxpayers/{id}/dashboard`). Throws an
/// `ApiFailure` when the server refuses or cannot be reached.
abstract interface class DashboardRepository {
  /// Home of [taxpayerId] for the current month in America/Mexico_City
  /// (no period: the server's clock decides), with the client's default
  /// 12-month trend (MODULE-dashboard OQ-D1).
  Future<Dashboard> getDashboard({required String taxpayerId});
}

/// [DashboardRepository] over the generated client.
class ApiDashboardRepository implements DashboardRepository {
  /// The repository on [_client].
  new(this._client);

  final DashboardClient _client;

  @override
  Future<Dashboard> getDashboard({required String taxpayerId}) =>
      guardApi(() => _client.getDashboard(taxpayerId: taxpayerId));
}

/// The dashboard read. Tests override it with a fake.
final dashboardRepositoryProvider = Provider<DashboardRepository>(
  (ref) => ApiDashboardRepository(ref.watch(apiClientProvider).dashboard),
  name: 'dashboardRepositoryProvider',
);
