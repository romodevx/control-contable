import 'package:control_contable/api/generated/clients/tax_engine_client.dart';
import 'package:control_contable/api/generated/clients/taxpayers_client.dart';
import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/regime.dart';
import 'package:control_contable/api/generated/models/tax_input_value.dart';
import 'package:control_contable/api/generated/models/tax_inputs_update.dart';
import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/api/generated/models/taxpayer_create.dart';
import 'package:control_contable/api/generated/models/taxpayer_regime.dart';
import 'package:control_contable/api/generated/models/taxpayer_regime_create.dart';
import 'package:control_contable/api/generated/models/taxpayer_regime_update.dart';
import 'package:control_contable/api/generated/models/taxpayer_update.dart';
import 'package:control_contable/core/network/api_client_provider.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The taxpayer operations of onboarding and settings (`/v1/regimes`,
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

  /// Every taxpayer of the user, archived ones too, by `created_at`
  /// (`GET /v1/taxpayers?include_archived=true`; at most 10, unpaginated).
  Future<List<Taxpayer>> listTaxpayers();

  /// Renames a taxpayer (`PATCH /v1/taxpayers/{id}`).
  Future<Taxpayer> rename({
    required String taxpayerId,
    required String displayName,
  });

  /// Archives a taxpayer (`POST …/archive`; allowed while read-only).
  Future<Taxpayer> archive(String taxpayerId);

  /// Unarchives a taxpayer (`POST …/unarchive`).
  Future<Taxpayer> unarchive(String taxpayerId);

  /// Deletes a taxpayer and every record of it (`DELETE …?confirm=true`).
  /// Irreversible: call it only after the user confirmed.
  Future<void> delete(String taxpayerId);

  /// Adds a regime, or reopens an ended one (`POST …/regimes`), starting
  /// in [startPeriod] (`YYYY-MM`; `null` is the current month). A
  /// reopened regime keeps its original start.
  Future<TaxpayerRegime> addRegime({
    required String taxpayerId,
    required String regimeCode,
    Map<String, Object>? options,
    String? startPeriod,
  });

  /// Replaces the options sent (`PATCH …/regimes/{code}`); the change
  /// applies to every period of the regime (OQ-7).
  Future<TaxpayerRegime> updateRegimeOptions({
    required String taxpayerId,
    required String regimeCode,
    required Map<String, Object> options,
  });

  /// Ends a regime at [endPeriod] (`DELETE …/regimes/{code}`); [confirm]
  /// keeps its records when it has some (`409 regime_has_records` without).
  Future<TaxpayerRegime> endRegime({
    required String taxpayerId,
    required String regimeCode,
    required String endPeriod,
    bool confirm = false,
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
  Future<List<Taxpayer>> listTaxpayers() async {
    final list = await guardApi(
      () => _taxpayers.listTaxpayers(includeArchived: true),
    );
    return list.items;
  }

  @override
  Future<Taxpayer> rename({
    required String taxpayerId,
    required String displayName,
  }) => guardApi(
    () => _taxpayers.updateTaxpayer(
      taxpayerId: taxpayerId,
      body: TaxpayerUpdate(displayName: displayName),
    ),
  );

  @override
  Future<Taxpayer> archive(String taxpayerId) =>
      guardApi(() => _taxpayers.archiveTaxpayer(taxpayerId: taxpayerId));

  @override
  Future<Taxpayer> unarchive(String taxpayerId) =>
      guardApi(() => _taxpayers.unarchiveTaxpayer(taxpayerId: taxpayerId));

  @override
  Future<void> delete(String taxpayerId) => guardApi(
    () => _taxpayers.deleteTaxpayer(taxpayerId: taxpayerId, confirm: true),
  );

  @override
  Future<TaxpayerRegime> addRegime({
    required String taxpayerId,
    required String regimeCode,
    Map<String, Object>? options,
    String? startPeriod,
  }) => guardApi(
    () => _taxpayers.addTaxpayerRegime(
      taxpayerId: taxpayerId,
      body: TaxpayerRegimeCreate(
        regimeCode: regimeCode,
        options: options,
        startPeriod: startPeriod,
      ),
    ),
  );

  @override
  Future<TaxpayerRegime> updateRegimeOptions({
    required String taxpayerId,
    required String regimeCode,
    required Map<String, Object> options,
  }) => guardApi(
    () => _taxpayers.updateTaxpayerRegime(
      taxpayerId: taxpayerId,
      regimeCode: regimeCode,
      body: TaxpayerRegimeUpdate(options: options),
    ),
  );

  @override
  Future<TaxpayerRegime> endRegime({
    required String taxpayerId,
    required String regimeCode,
    required String endPeriod,
    bool confirm = false,
  }) => guardApi(
    () => _taxpayers.archiveTaxpayerRegime(
      taxpayerId: taxpayerId,
      regimeCode: regimeCode,
      endPeriod: endPeriod,
      confirm: confirm,
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
