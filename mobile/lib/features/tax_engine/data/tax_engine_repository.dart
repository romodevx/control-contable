import 'package:control_contable/api/generated/clients/tax_engine_client.dart';
import 'package:control_contable/api/generated/models/breakdown.dart';
import 'package:control_contable/api/generated/models/tax_input_value.dart';
import 'package:control_contable/api/generated/models/tax_inputs.dart';
import 'package:control_contable/api/generated/models/tax_inputs_update.dart';
import 'package:control_contable/api/generated/models/taxpayer_estimate.dart';
import 'package:control_contable/core/network/api_client_provider.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The tax engine operations of M5 (`/v1/taxpayers/{id}/estimates/**` and
/// `…/regimes/{code}/tax-inputs`). Every method throws an `ApiFailure` when
/// the server refuses or cannot be reached.
abstract interface class TaxEngineRepository {
  /// The informational estimate of [taxpayerId] for [period] (`YYYY-MM`),
  /// or for the current month in America/Mexico_City when [period] is
  /// `null` (the server's clock decides; `GET …/estimates`).
  Future<TaxpayerEstimate> getEstimate({
    required String taxpayerId,
    String? period,
  });

  /// The step-by-step computation of [regime] in [period] (`YYYY-MM`),
  /// with its legal basis (`GET …/estimates/{period}/breakdown`).
  Future<Breakdown> getBreakdown({
    required String taxpayerId,
    required String period,
    required String regime,
  });

  /// The tax inputs of [regime] for [fiscalYear], with their definitions.
  Future<TaxInputs> getTaxInputs({
    required String taxpayerId,
    required String regime,
    required int fiscalYear,
  });

  /// Replaces every tax input of [regime] for [fiscalYear] with [values];
  /// answers what is stored. Idempotent.
  Future<TaxInputs> putTaxInputs({
    required String taxpayerId,
    required String regime,
    required int fiscalYear,
    required List<TaxInputValue> values,
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

  @override
  Future<Breakdown> getBreakdown({
    required String taxpayerId,
    required String period,
    required String regime,
  }) => guardApi(
    () => _client.getEstimateBreakdown(
      taxpayerId: taxpayerId,
      period: period,
      regime: regime,
    ),
  );

  @override
  Future<TaxInputs> getTaxInputs({
    required String taxpayerId,
    required String regime,
    required int fiscalYear,
  }) => guardApi(
    () => _client.getTaxInputs(
      taxpayerId: taxpayerId,
      regimeCode: regime,
      fiscalYear: fiscalYear,
    ),
  );

  @override
  Future<TaxInputs> putTaxInputs({
    required String taxpayerId,
    required String regime,
    required int fiscalYear,
    required List<TaxInputValue> values,
  }) => guardApi(
    () => _client.putTaxInputs(
      taxpayerId: taxpayerId,
      regimeCode: regime,
      fiscalYear: fiscalYear,
      body: TaxInputsUpdate(values: values),
    ),
  );
}

/// The tax engine operations. Tests override them with a fake.
final taxEngineRepositoryProvider = Provider<TaxEngineRepository>(
  (ref) => ApiTaxEngineRepository(ref.watch(apiClientProvider).taxEngine),
  name: 'taxEngineRepositoryProvider',
);
