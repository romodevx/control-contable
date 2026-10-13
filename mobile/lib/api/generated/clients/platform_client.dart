// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

import '../models/enabled_flags.dart';
import '../models/health_status.dart';

part 'platform_client.g.dart';

@RestApi()
abstract class PlatformClient {
  factory PlatformClient(Dio dio, {String? baseUrl}) = _PlatformClient;

  /// Liveness and readiness
  @GET('/healthz')
  Future<HealthStatus> getHealth();

  /// Modules enabled for the client (UX only; server enforces)
  @GET('/v1/flags')
  Future<EnabledFlags> listEnabledFlags();
}
