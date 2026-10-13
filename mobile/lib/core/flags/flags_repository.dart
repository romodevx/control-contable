import 'package:control_contable/api/generated/clients/platform_client.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/network/api_client_provider.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Source of the modules the server has enabled for this client.
abstract interface class FlagsRepository {
  /// The enabled modules. Keys this build does not know are dropped.
  /// Throws when the server cannot be asked.
  Future<Set<AppModule>> fetchEnabled();
}

/// [FlagsRepository] over `GET /v1/flags` (the generated platform client).
class ApiFlagsRepository implements FlagsRepository {
  /// Creates the repository on the generated platform client.
  new(this._client);

  final PlatformClient _client;

  @override
  Future<Set<AppModule>> fetchEnabled() async {
    final response = await _client.listEnabledFlags();
    return {for (final key in response.enabled) ?AppModule.fromKey(key)};
  }
}

/// The flags source. Tests override it with a fake.
final flagsRepositoryProvider = Provider<FlagsRepository>(
  (ref) => ApiFlagsRepository(ref.watch(apiClientProvider).platform),
  name: 'flagsRepositoryProvider',
);
