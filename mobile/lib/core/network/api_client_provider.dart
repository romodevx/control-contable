import 'package:control_contable/api/generated/api_client.dart';
import 'package:control_contable/core/network/dio_provider.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The API client generated from `docs/api/openapi.yaml` (`make mobile-client`,
/// ADR-0003). It shares the one [dioProvider] instance, so the per-flavor
/// base URL, timeouts, HTTPS guard and (later) auth interceptors apply to
/// every call. Features read their typed client from it, e.g.
/// `ref.watch(apiClientProvider).platform`.
final apiClientProvider = Provider<ApiClient>(
  (ref) => ApiClient(ref.watch(dioProvider)),
  name: 'apiClientProvider',
);
