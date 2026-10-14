import 'package:control_contable/api/generated/api_client.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/tax_engine/data/tax_engine_repository.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/recording_adapter.dart';
import 'estimate_fixtures.dart';

const String _path = '/v1/taxpayers/$fixtureTaxpayerId/estimates';

void main() {
  late RecordingAdapter adapter;
  late ApiTaxEngineRepository repository;

  setUp(() {
    adapter = RecordingAdapter(
      replies: {_path: (status: 200, body: completeEstimateJson)},
    );
    final dio = Dio(BaseOptions(baseUrl: 'http://localhost:3001'))
      ..httpClientAdapter = adapter;
    repository = ApiTaxEngineRepository(ApiClient(dio).taxEngine);
  });

  test('the current month is asked without a period (server clock)', () async {
    final estimate = await repository.getEstimate(
      taxpayerId: fixtureTaxpayerId,
    );

    final request = adapter.requests.single;
    expect(request.method, 'GET');
    expect(request.uri.path, _path);
    expect(request.uri.queryParameters, isNot(contains('period')));
    expect(estimate.totals.totalPayableRoundedCents, 228429);
    expect(estimate.regimes.single.regimeCode, '612');
  });

  test('a past month is asked with its period', () async {
    await repository.getEstimate(
      taxpayerId: fixtureTaxpayerId,
      period: '2026-02',
    );

    expect(adapter.requests.single.uri.queryParameters['period'], '2026-02');
  });

  test('a refusal surfaces as an ApiFailure', () async {
    adapter.reply(_path, (
      status: 422,
      body: {
        'error': {
          'code': 'business_rule_violation',
          'message': 'no params',
          'details': [
            {
              'field': 'period',
              'code': 'parameters_unavailable',
              'message': 'x',
            },
          ],
        },
      },
    ));

    final failure = await failureOf(
      () => repository.getEstimate(taxpayerId: fixtureTaxpayerId),
    );

    expect(failure, isA<ApiRejected>());
  });
}
