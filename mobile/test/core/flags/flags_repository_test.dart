import 'package:control_contable/api/generated/clients/platform_client.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flags_repository.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/recording_adapter.dart';

const _flagsPath = '/v1/flags';

FlagsRepository _repositoryAnswering(CannedReply reply) {
  final dio = Dio(BaseOptions(baseUrl: 'http://api.test'))
    ..httpClientAdapter = RecordingAdapter(replies: {_flagsPath: reply});
  return ApiFlagsRepository(PlatformClient(dio));
}

void main() {
  test('maps GET /v1/flags keys to modules', () async {
    final repository = _repositoryAnswering((
      status: 200,
      body: {
        'enabled': ['module_auth', 'module_tax_engine'],
      },
    ));

    expect(await repository.fetchEnabled(), {
      AppModule.auth,
      AppModule.taxEngine,
    });
  });

  test('drops keys this build does not know', () async {
    final repository = _repositoryAnswering((
      status: 200,
      body: {
        'enabled': ['module_from_the_future', 'module_dashboard'],
      },
    ));

    expect(await repository.fetchEnabled(), {AppModule.dashboard});
  });

  test('a server error propagates (the provider decides: fail-closed)', () {
    final repository = _repositoryAnswering((
      status: 500,
      body: <String, Object?>{},
    ));

    expect(repository.fetchEnabled(), throwsA(isA<DioException>()));
  });

  test('a malformed body propagates instead of enabling anything', () {
    final repository = _repositoryAnswering((
      status: 200,
      body: {'enabled': 'module_auth'},
    ));

    expect(repository.fetchEnabled(), throwsA(anything));
  });
}
