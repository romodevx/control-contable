// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

import '../models/dashboard.dart';
import '../models/period.dart';
import '../models/uuid.dart';

part 'dashboard_client.g.dart';

@RestApi()
abstract class DashboardClient {
  factory DashboardClient(Dio dio, {String? baseUrl}) = _DashboardClient;

  /// Home for the active taxpayer
  @GET('/v1/taxpayers/{taxpayerId}/dashboard')
  Future<Dashboard> getDashboard({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Query('period') required Period period,
    @Query('months') int? months = 12,
  });
}
