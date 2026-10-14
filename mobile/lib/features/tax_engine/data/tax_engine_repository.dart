import 'package:control_contable/api/generated/clients/tax_engine_client.dart';
import 'package:control_contable/api/generated/models/taxpayer_estimate.dart';
import 'package:control_contable/core/network/api_client_provider.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The tax engine reads of M5 (`/v1/taxpayers/{id}/estimates/**`). Every
/// method throws an `ApiFailure` when the server refuses or cannot be
/// reached.
abstract interface class TaxEngineRepository {
  /// The informational estimate of [taxpayerId] for [period] (`YYYY-MM`),
  /// or for the current month in America/Mexico_City when [period] is
  /// `null` (the server's clock decides; `GET …/estimates`).
  Future<TaxpayerEstimate> getEstimate({
    required String taxpayerId,
    String? period,
  });
}

/// [TaxEngineRepository] over the generated client.
class ApiTaxEngineRepository implements TaxEngineRepository {
  /// The repository on [_client].
  new(this._client);

  final TaxEngineClient _client;

  @override
  Future<TaxpayerEstimate> getEstimate({
    required String taxpayerId,
    String? period,
  }) => guardApi(
    () => _client.getTaxpayerEstimate(taxpayerId: taxpayerId, period: period),
  );
}

/// The tax engine reads. Tests override it with a fake.
final taxEngineRepositoryProvider = Provider<TaxEngineRepository>(
  (ref) => ApiTaxEngineRepository(ref.watch(apiClientProvider).taxEngine),
  name: 'taxEngineRepositoryProvider',
);
