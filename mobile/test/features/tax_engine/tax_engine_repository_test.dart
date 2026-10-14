import 'dart:convert';

import 'package:control_contable/api/generated/api_client.dart';
import 'package:control_contable/api/generated/models/tax_input_value.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/tax_engine/data/tax_engine_repository.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/recording_adapter.dart';
import 'breakdown_fixtures.dart';
import 'estimate_fixtures.dart';
import 'tax_inputs_fixtures.dart';

const String _path = '/v1/taxpayers/$fixtureTaxpayerId/estimates';
const String _breakdownPath = '$_path/2026-03/breakdown';
const String _inputsPath =
    '/v1/taxpayers/$fixtureTaxpayerId/regimes/612/tax-inputs';

void main() {
  late RecordingAdapter adapter;
  late ApiTaxEngineRepository repository;

  setUp(() {
    adapter = RecordingAdapter(
      replies: {
        _path: (status: 200, body: completeEstimateJson),
        _breakdownPath: (status: 200, body: breakdownJson()),
        _inputsPath: (status: 200, body: taxInputsJson()),
      },
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

  test('a breakdown is asked by month and regime', () async {
    final breakdown = await repository.getBreakdown(
      taxpayerId: fixtureTaxpayerId,
      period: '2026-03',
      regime: '612',
    );

    final request = adapter.requests.single;
    expect(request.method, 'GET');
    expect(request.uri.path, _breakdownPath);
    expect(request.uri.queryParameters, {'regime': '612'});
    expect(
      breakdown.lines.first.concept,
      'Ingresos cobrados acumulados ene–mar',
    );
  });

  test('tax inputs are read by regime and fiscal year', () async {
    final inputs = await repository.getTaxInputs(
      taxpayerId: fixtureTaxpayerId,
      regime: '612',
      fiscalYear: 2026,
    );

    final request = adapter.requests.single;
    expect(request.method, 'GET');
    expect(request.uri.path, _inputsPath);
    expect(request.uri.queryParameters, {'fiscal_year': '2026'});
    expect(inputs.values.single.valueCents, 200000);
  });

  test(
    'a save replaces the values, sending only the field of each kind',
    () async {
      await repository.putTaxInputs(
        taxpayerId: fixtureTaxpayerId,
        regime: '612',
        fiscalYear: 2026,
        values: const [
          TaxInputValue(
            key: 'prior_provisional_payments',
            month: 2,
            valueCents: 150000,
          ),
          TaxInputValue(key: 'coeficiente_utilidad', valueRate: '0.15'),
        ],
      );

      final request = adapter.requests.single;
      expect(request.method, 'PUT');
      expect(request.uri.queryParameters, {'fiscal_year': '2026'});
      expect(jsonDecode(jsonEncode(request.data)), {
        'values': [
          {
            'key': 'prior_provisional_payments',
            'month': 2,
            'value_cents': 150000,
          },
          {'key': 'coeficiente_utilidad', 'value_rate': '0.15'},
        ],
      });
    },
  );
}
