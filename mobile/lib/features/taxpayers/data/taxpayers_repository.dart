import 'package:control_contable/api/generated/clients/tax_engine_client.dart';
import 'package:control_contable/api/generated/clients/taxpayers_client.dart';
import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/regime.dart';
import 'package:control_contable/api/generated/models/tax_input_value.dart';
import 'package:control_contable/api/generated/models/tax_inputs_update.dart';
import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/api/generated/models/taxpayer_create.dart';
import 'package:control_contable/api/generated/models/taxpayer_update.dart';
import 'package:control_contable/core/network/api_client_provider.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The taxpayer operations onboarding needs (`/v1/regimes`,
/// `/v1/taxpayers/**`, and M5's tax inputs). Every method throws an
/// `ApiFailure` when the server refuses or cannot be reached.
abstract interface class TaxpayersRepository {
  /// The regime catalog for [personType] (`GET /v1/regimes`).
  Future<List<Regime>> listRegimes(PersonType personType);

  /// Creates the taxpayer with its regimes and options in one transaction
  /// (`POST /v1/taxpayers`). Retrying with the same client `id` answers the
  /// existing taxpayer.
  Future<Taxpayer> createTaxpayer(TaxpayerCreate request);

  /// Corrects a mistyped RFC of the same person type
  /// (`PATCH /v1/taxpayers/{id}`, OQ-1).
  Future<Taxpayer> correctRfc({
    required String taxpayerId,
    required String rfc,
  });

  /// Replaces one regime's tax inputs for [fiscalYear]
  /// (`PUT …/regimes/{code}/tax-inputs`, M5).
  Future<void> saveTaxInputs({
    required String taxpayerId,
    required String regimeCode,
    required int fiscalYear,
    required List<TaxInputValue> values,
  });
}

/// [TaxpayersRepository] over the generated clients.
class ApiTaxpayersRepository implements TaxpayersRepository {
  /// The repository on [_taxpayers] and [_taxEngine].
  new(this._taxpayers, this._taxEngine);

  final TaxpayersClient _taxpayers;
  final TaxEngineClient _taxEngine;

  @override
  Future<List<Regime>> listRegimes(PersonType personType) async {
    final list = await guardApi(
      () => _taxpayers.listRegimes(personType: personType),
    );
    return list.items;
  }

  @override
  Future<Taxpayer> createTaxpayer(TaxpayerCreate request) =>
      guardApi(() => _taxpayers.createTaxpayer(body: request));

  @override
  Future<Taxpayer> correctRfc({
    required String taxpayerId,
    required String rfc,
  }) => guardApi(
    () => _taxpayers.updateTaxpayer(
      taxpayerId: taxpayerId,
      body: TaxpayerUpdate(rfc: rfc),
    ),
  );

  @override
  Future<void> saveTaxInputs({
    required String taxpayerId,
    required String regimeCode,
    required int fiscalYear,
    required List<TaxInputValue> values,
  }) => guardApi(
    () => _taxEngine.putTaxInputs(
      taxpayerId: taxpayerId,
      regimeCode: regimeCode,
      fiscalYear: fiscalYear,
      body: TaxInputsUpdate(values: values),
    ),
  );
}

/// The taxpayer operations. Tests override it with a fake.
final taxpayersRepositoryProvider = Provider<TaxpayersRepository>((ref) {
  final api = ref.watch(apiClientProvider);
  return ApiTaxpayersRepository(api.taxpayers, api.taxEngine);
}, name: 'taxpayersRepositoryProvider');
