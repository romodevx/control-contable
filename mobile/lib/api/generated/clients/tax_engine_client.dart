// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

import '../models/breakdown.dart';
import '../models/fiscal_year.dart';
import '../models/period.dart';
import '../models/regime_code.dart';
import '../models/tax_inputs.dart';
import '../models/tax_inputs_update.dart';
import '../models/taxpayer_estimate.dart';
import '../models/uuid.dart';

part 'tax_engine_client.g.dart';

@RestApi()
abstract class TaxEngineClient {
  factory TaxEngineClient(Dio dio, {String? baseUrl}) = _TaxEngineClient;

  /// Monthly IVA + provisional ISR estimate, per regime and total
  @GET('/v1/taxpayers/{taxpayerId}/estimates')
  Future<TaxpayerEstimate> getTaxpayerEstimate({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Query('period') required Period period,
  });

  /// Step-by-step breakdown with legal basis for one regime
  @GET('/v1/taxpayers/{taxpayerId}/estimates/{period}/breakdown')
  Future<Breakdown> getEstimateBreakdown({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('period') required Period period,
    @Query('regime') required RegimeCode regime,
  });

  /// Tax inputs (YTD, coeficiente, losses, prior payments) with which keys are required
  @GET('/v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs')
  Future<TaxInputs> getTaxInputs({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('regimeCode') required RegimeCode regimeCode,
    @Query('fiscal_year') required FiscalYear fiscalYear,
  });

  /// Replace tax inputs for the fiscal year
  @PUT('/v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs')
  Future<TaxInputs> putTaxInputs({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('regimeCode') required RegimeCode regimeCode,
    @Query('fiscal_year') required FiscalYear fiscalYear,
    @Body() required TaxInputsUpdate body,
  });
}
