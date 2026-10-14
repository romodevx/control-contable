import 'dart:convert';

import 'package:control_contable/api/generated/api_client.dart';
import 'package:control_contable/api/generated/models/history_viewed_parameter_set.dart';
import 'package:control_contable/api/generated/models/history_viewed_request.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/history/data/history_repository.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/recording_adapter.dart';
import '../tax_engine/estimate_fixtures.dart';
import 'history_fixtures.dart';

const String _path = '/v1/taxpayers/$fixtureTaxpayerId/history';
const String _viewedPath = '$_path/2026-02/viewed';

const HistoryViewedRequest _request = HistoryViewedRequest(
  recordsFingerprint: historyRecordsFingerprint,
  figuresFingerprint: fixtureFiguresFingerprint,
  engineVersion: '1.0.0',
  parameterSet: HistoryViewedParameterSet(fiscalYear: 2026, version: 1),
);

void main() {
  late RecordingAdapter adapter;
  late ApiHistoryRepository repository;

  setUp(() {
    adapter = RecordingAdapter(
      replies: {
        _path: (status: 200, body: historyYearJson()),
        _viewedPath: (status: 204, body: null),
      },
    );
    final dio = Dio(BaseOptions(baseUrl: 'http://localhost:3001'))
      ..httpClientAdapter = adapter;
    repository = ApiHistoryRepository(ApiClient(dio).history);
  });

  test('reads one fiscal year of the taxpayer', () async {
    final year = await repository.getHistory(
      taxpayerId: fixtureTaxpayerId,
      fiscalYear: 2026,
    );

    final request = adapter.requests.single;
    expect(request.method, 'GET');
    expect(request.uri.path, _path);
    expect(request.uri.queryParameters, {'fiscal_year': '2026'});
    expect(year.months.map((m) => m.period), ['2026-03', '2026-02', '2026-01']);
  });

  test('posts what the app rendered as the closed body', () async {
    await repository.markViewed(
      taxpayerId: fixtureTaxpayerId,
      period: '2026-02',
      request: _request,
    );

    final request = adapter.requests.single;
    expect(request.method, 'POST');
    expect(request.uri.path, _viewedPath);
    expect(jsonDecode(jsonEncode(request.data)), {
      'records_fingerprint': historyRecordsFingerprint,
      'figures_fingerprint': fixtureFiguresFingerprint,
      'engine_version': '1.0.0',
      'parameter_set': {'fiscal_year': 2026, 'version': 1},
    });
  });

  test('a 404 (module OFF, foreign taxpayer) is an ApiNotFound', () async {
    adapter.reply(_path, (
      status: 404,
      body: {
        'error': {'code': 'not_found', 'message': 'x'},
      },
    ));

    await expectLater(
      repository.getHistory(taxpayerId: fixtureTaxpayerId, fiscalYear: 2026),
      throwsA(isA<ApiNotFound>()),
    );
  });
}
