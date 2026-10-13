// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

import '../models/fiscal_year.dart';
import '../models/history_year.dart';
import '../models/period.dart';
import '../models/uuid.dart';

part 'history_client.g.dart';

@RestApi()
abstract class HistoryClient {
  factory HistoryClient(Dio dio, {String? baseUrl}) = _HistoryClient;

  /// Monthly summaries for a fiscal year
  @GET('/v1/taxpayers/{taxpayerId}/history')
  Future<HistoryYear> getHistory({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Query('fiscal_year') required FiscalYear fiscalYear,
  });

  /// Mark a month as viewed (stores the current records fingerprint)
  @POST('/v1/taxpayers/{taxpayerId}/history/{period}/viewed')
  Future<void> markHistoryViewed({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('period') required Period period,
  });
}
