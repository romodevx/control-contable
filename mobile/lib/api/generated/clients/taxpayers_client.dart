// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

import '../models/period.dart';
import '../models/person_type.dart';
import '../models/regime_code.dart';
import '../models/regime_list.dart';
import '../models/taxpayer.dart';
import '../models/taxpayer_create.dart';
import '../models/taxpayer_list.dart';
import '../models/taxpayer_regime.dart';
import '../models/taxpayer_regime_create.dart';
import '../models/taxpayer_regime_update.dart';
import '../models/taxpayer_update.dart';
import '../models/uuid.dart';

part 'taxpayers_client.g.dart';

@RestApi()
abstract class TaxpayersClient {
  factory TaxpayersClient(Dio dio, {String? baseUrl}) = _TaxpayersClient;

  /// Regime catalog filtered by person type
  @GET('/v1/regimes')
  Future<RegimeList> listRegimes({
    @Query('person_type') required PersonType personType,
  });

  /// Taxpayers managed by the user
  @GET('/v1/taxpayers')
  Future<TaxpayerList> listTaxpayers({
    @Query('include_archived') bool? includeArchived = false,
  });

  /// Create a taxpayer (onboarding order - person type, RFC, regimes, options)
  @POST('/v1/taxpayers')
  Future<Taxpayer> createTaxpayer({@Body() required TaxpayerCreate body});

  /// Get a taxpayer
  @GET('/v1/taxpayers/{taxpayerId}')
  Future<Taxpayer> getTaxpayer({@Path('taxpayerId') required Uuid taxpayerId});

  /// Rename a taxpayer (person type and RFC are immutable)
  @PATCH('/v1/taxpayers/{taxpayerId}')
  Future<Taxpayer> updateTaxpayer({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Body() required TaxpayerUpdate body,
  });

  /// Permanently delete a taxpayer and its records (irreversible).
  ///
  /// [confirm] - Must be `true` to perform a destructive or data-affecting action.
  @DELETE('/v1/taxpayers/{taxpayerId}')
  Future<void> deleteTaxpayer({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Query('confirm') bool? confirm = false,
  });

  /// Archive (hide) a taxpayer, keeping its data
  @POST('/v1/taxpayers/{taxpayerId}/archive')
  Future<Taxpayer> archiveTaxpayer({
    @Path('taxpayerId') required Uuid taxpayerId,
  });

  /// Add a regime to a taxpayer
  @POST('/v1/taxpayers/{taxpayerId}/regimes')
  Future<TaxpayerRegime> addTaxpayerRegime({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Body() required TaxpayerRegimeCreate body,
  });

  /// Update regime options
  @PATCH('/v1/taxpayers/{taxpayerId}/regimes/{regimeCode}')
  Future<TaxpayerRegime> updateTaxpayerRegime({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('regimeCode') required RegimeCode regimeCode,
    @Body() required TaxpayerRegimeUpdate body,
  });

  /// Archive a regime with an end period (records are kept).
  ///
  /// If the regime has records and `confirm` is not `true`, answers `409 regime_has_records`.
  ///
  /// [confirm] - Must be `true` to perform a destructive or data-affecting action.
  @DELETE('/v1/taxpayers/{taxpayerId}/regimes/{regimeCode}')
  Future<TaxpayerRegime> archiveTaxpayerRegime({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('regimeCode') required RegimeCode regimeCode,
    @Query('end_period') required Period endPeriod,
    @Query('confirm') bool? confirm = false,
  });
}
