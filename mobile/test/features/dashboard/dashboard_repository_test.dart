import 'package:control_contable/api/generated/api_client.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/dashboard/data/dashboard_repository.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/recording_adapter.dart';
import '../tax_engine/estimate_fixtures.dart';
import 'dashboard_fixtures.dart';

const String _path = '/v1/taxpayers/$fixtureTaxpayerId/dashboard';

void main() {
  late RecordingAdapter adapter;
  late ApiDashboardRepository repository;

  setUp(() {
    adapter = RecordingAdapter(
      replies: {_path: (status: 200, body: dashboardJson())},
    );
    final dio = Dio(BaseOptions(baseUrl: 'http://localhost:3001'))
      ..httpClientAdapter = adapter;
    repository = ApiDashboardRepository(ApiClient(dio).dashboard);
  });

  test('Home asks for 12 months and no period (server clock)', () async {
    final dashboard = await repository.getDashboard(
      taxpayerId: fixtureTaxpayerId,
    );

    final request = adapter.requests.single;
    expect(request.method, 'GET');
    expect(request.uri.path, _path);
    expect(request.uri.queryParameters, {'months': '12'});
    expect(dashboard.estimate?.totals.totalPayableRoundedCents, 228429);
    expect(dashboard.trend, hasLength(3));
  });

  test('a 404 (module OFF, foreign taxpayer) is an ApiNotFound', () async {
    adapter.reply(_path, (
      status: 404,
      body: {
        'error': {'code': 'not_found', 'message': 'x'},
      },
    ));

    await expectLater(
      repository.getDashboard(taxpayerId: fixtureTaxpayerId),
      throwsA(isA<ApiNotFound>()),
    );
  });
}
