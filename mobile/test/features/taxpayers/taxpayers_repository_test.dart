import 'dart:convert';

import 'package:control_contable/api/generated/api_client.dart';
import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/tax_input_value.dart';
import 'package:control_contable/api/generated/models/taxpayer_create.dart';
import 'package:control_contable/api/generated/models/taxpayer_regime_create.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/taxpayers/data/taxpayers_repository.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/recording_adapter.dart';

const String _taxpayerId = '0192f0c4-7b1e-7c3a-9d2e-6f1a2b3c4d5e';

const Map<String, Object?> _taxpayer = {
  'id': _taxpayerId,
  'person_type': 'fisica',
  'display_name': 'Ana',
  'rfc_masked': 'GOMA******AB6',
  'regimes': <Object?>[],
  'archived_at': null,
  'created_at': '2026-10-13T16:00:00Z',
  'warnings': [
    {'code': 'rfc_check_digit_mismatch', 'field': 'rfc', 'message': 'Revisa.'},
  ],
};

const Map<String, Object?> _regimes = {
  'items': [
    {
      'code': '606',
      'name': 'Arrendamiento',
      'person_type': 'fisica',
      'treatment': 'full',
      'explanation': null,
      'option_defs': [
        {
          'key': 'deduction_mode',
          'label': 'Forma de deducir',
          'type': 'enum',
          'required': true,
          'allowed_values': ['blind_35', 'actual'],
          'default': 'blind_35',
        },
      ],
      'required_inputs': <Object?>[],
      'incompatible_with': <Object?>[],
    },
  ],
};

void main() {
  late RecordingAdapter adapter;
  late ApiTaxpayersRepository repository;

  setUp(() {
    adapter = RecordingAdapter(
      replies: {
        '/v1/regimes': (status: 200, body: _regimes),
        '/v1/taxpayers': (status: 201, body: _taxpayer),
        '/v1/taxpayers/$_taxpayerId': (status: 200, body: _taxpayer),
        '/v1/taxpayers/$_taxpayerId/regimes/601/tax-inputs': (
          status: 200,
          body: {
            'regime_code': '601',
            'fiscal_year': 2026,
            'definitions': <Object?>[],
            'values': <Object?>[],
          },
        ),
      },
    );
    final dio = Dio(BaseOptions(baseUrl: 'http://localhost:3001'))
      ..httpClientAdapter = adapter;
    final api = ApiClient(dio);
    repository = ApiTaxpayersRepository(api.taxpayers, api.taxEngine);
  });

  Map<String, Object?> lastBody() =>
      jsonDecode(jsonEncode(adapter.requests.last.data))
          as Map<String, Object?>;

  test('listRegimes asks for one person type and reads its options', () async {
    final regimes = await repository.listRegimes(PersonType.moral);

    final request = adapter.requests.single;
    expect(request.method, 'GET');
    expect(request.uri.queryParameters, {'person_type': 'moral'});
    expect(regimes.single.optionDefs.single.defaultValue, 'blind_35');
  });

  test('createTaxpayer posts the whole onboarding in one call', () async {
    final created = await repository.createTaxpayer(
      const TaxpayerCreate(
        id: _taxpayerId,
        personType: PersonType.fisica,
        rfc: 'GOMA800101AB6',
        displayName: 'Ana',
        regimes: [
          TaxpayerRegimeCreate(
            regimeCode: '606',
            options: {'deduction_mode': 'actual'},
          ),
        ],
      ),
    );

    expect(adapter.requests.single.method, 'POST');
    expect(lastBody(), {
      'id': _taxpayerId,
      'person_type': 'fisica',
      'rfc': 'GOMA800101AB6',
      'display_name': 'Ana',
      'regimes': [
        {
          'regime_code': '606',
          'options': {'deduction_mode': 'actual'},
        },
      ],
    });
    expect(created.rfcMasked, 'GOMA******AB6');
    expect(created.warnings?.single.code, 'rfc_check_digit_mismatch');
  });

  test('correctRfc patches only the RFC', () async {
    await repository.correctRfc(taxpayerId: _taxpayerId, rfc: 'GOMA800101AB7');

    expect(adapter.requests.single.method, 'PATCH');
    expect(lastBody(), {'rfc': 'GOMA800101AB7'});
  });

  test('saveTaxInputs puts the values for the fiscal year', () async {
    await repository.saveTaxInputs(
      taxpayerId: _taxpayerId,
      regimeCode: '601',
      fiscalYear: 2026,
      values: const [
        TaxInputValue(key: 'profit_coefficient', valueRate: '0.2'),
      ],
    );

    final request = adapter.requests.single;
    expect(request.method, 'PUT');
    expect(request.uri.queryParameters, {'fiscal_year': '2026'});
    expect(lastBody(), {
      'values': [
        {'key': 'profit_coefficient', 'value_rate': '0.2'},
      ],
    });
  });

  test('a 402 is an ApiEntitlementRequired', () async {
    adapter.reply('/v1/taxpayers', (
      status: 402,
      body: {
        'error': {
          'code': 'entitlement_required',
          'message': 'Read-only.',
          'request_id': 'req_1',
        },
      },
    ));

    final failure = await failureOf(
      () => repository.createTaxpayer(
        const TaxpayerCreate(
          personType: PersonType.fisica,
          rfc: 'GOMA800101AB6',
          displayName: 'Ana',
          regimes: [TaxpayerRegimeCreate(regimeCode: '626')],
        ),
      ),
    );

    expect(failure, isA<ApiEntitlementRequired>());
  });
}
