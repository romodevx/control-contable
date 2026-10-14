import 'dart:convert';

import 'package:control_contable/api/generated/api_client.dart';
import 'package:control_contable/api/generated/models/counterparty_type.dart';
import 'package:control_contable/api/generated/models/income_create.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/api/generated/models/platform_activity.dart';
import 'package:control_contable/api/generated/models/warning_code.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/incomes/data/incomes_repository.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/recording_adapter.dart';

const String _taxpayerId = '0192f0c4-7b1e-7c3a-9d2e-6f1a2b3c4d5e';
const String _incomeId = '0192f0c4-0000-7000-8000-000000000001';
const String _incomes = '/v1/taxpayers/$_taxpayerId/incomes';

const Map<String, Object?> _income = {
  'id': _incomeId,
  'taxpayer_id': _taxpayerId,
  'regime_code': '612',
  'counterparty_type': 'persona_moral',
  'amount_cents': 1000000,
  'iva_rate': 'rate_16',
  'iva_cents': 160000,
  'isr_withheld_cents': 100000,
  'iva_withheld_cents': 106667,
  'invoiced_on': null,
  'collected_on': '2026-10-05',
  'platform_activity': null,
  'cfdi_uuid': null,
  'description': 'Diseño de logotipo',
  'pending_collection': false,
  'warnings': [
    {
      'code': 'duplicate_cfdi_uuid',
      'field': 'cfdi_uuid',
      'message': 'Otro ingreso tiene este folio.',
    },
  ],
  'created_at': '2026-10-13T16:00:00Z',
  'updated_at': '2026-10-13T16:00:00Z',
};

const Map<String, Object?> _totals = {
  'count': 1,
  'amount_cents': 1000000,
  'iva_cents': 160000,
  'isr_withheld_cents': 100000,
  'iva_withheld_cents': 106667,
};

void main() {
  late RecordingAdapter adapter;
  late ApiIncomesRepository repository;

  setUp(() {
    adapter = RecordingAdapter(
      replies: {
        _incomes: (
          status: 200,
          body: {
            'items': [_income],
            'next_cursor': 'c2',
            'totals': _totals,
          },
        ),
        '$_incomes/defaults': (
          status: 200,
          body: {
            'iva_rates_allowed': ['rate_16', 'rate_0', 'exempt'],
            'suggested_iva_rate': 'rate_16',
            'isr_withholding_allowed': true,
            'isr_withholding_rate': '0.04',
            'iva_withholding_allowed': true,
            'iva_withholding_rate': '0.5',
            'platform_activity': 'lodging',
          },
        ),
        '$_incomes/$_incomeId': (status: 204, body: null),
        '$_incomes/$_incomeId/restore': (status: 200, body: _income),
      },
    );
    final dio = Dio(BaseOptions(baseUrl: 'http://localhost:3001'))
      ..httpClientAdapter = adapter;
    repository = ApiIncomesRepository(ApiClient(dio).incomes);
  });

  test(
    'list asks for one month (and regime, cursor) and reads the page',
    () async {
      final page = await repository.list(
        taxpayerId: _taxpayerId,
        period: '2026-10',
        regime: '612',
        cursor: 'c1',
      );

      final request = adapter.requests.single;
      expect(request.method, 'GET');
      expect(request.uri.queryParameters, {
        'period': '2026-10',
        'limit': '50',
        'regime': '612',
        'cursor': 'c1',
      });
      expect(page.items.single.collectedOn, '2026-10-05');
      expect(page.nextCursor, 'c2');
      expect(page.totals.ivaWithheldCents, 106667);
    },
  );

  test('create posts integer centavos and YYYY-MM-DD dates', () async {
    adapter.reply(_incomes, (status: 201, body: _income));

    final created = await repository.create(
      taxpayerId: _taxpayerId,
      income: const IncomeCreate(
        id: _incomeId,
        regimeCode: '612',
        counterpartyType: CounterpartyType.personaMoral,
        amountCents: 1000000,
        ivaRate: IvaRate.rate16,
        ivaCents: 160000,
        isrWithheldCents: 100000,
        collectedOn: '2026-10-05',
      ),
    );

    expect(adapter.requests.single.method, 'POST');
    expect(jsonDecode(jsonEncode(adapter.requests.single.data)), {
      'id': _incomeId,
      'regime_code': '612',
      'counterparty_type': 'persona_moral',
      'amount_cents': 1000000,
      'iva_rate': 'rate_16',
      'iva_cents': 160000,
      'isr_withheld_cents': 100000,
      'collected_on': '2026-10-05',
    });
    expect(created.warnings.single.code, WarningCode.duplicateCfdiUuid);
  });

  test('defaults ask for the regime, payer, date and activity', () async {
    final defaults = await repository.defaults(
      taxpayerId: _taxpayerId,
      regime: '625',
      counterparty: CounterpartyType.plataforma,
      on: '2026-10-13',
      platformActivity: PlatformActivity.lodging,
    );

    expect(adapter.requests.single.uri.queryParameters, {
      'regime': '625',
      'counterparty_type': 'plataforma',
      'on': '2026-10-13',
      'platform_activity': 'lodging',
    });
    expect(defaults.isrWithholdingRate, '0.04');
    expect(defaults.ivaRatesAllowed, isNot(contains(IvaRate.rate8)));
  });

  test('delete and restore address one income', () async {
    await repository.delete(taxpayerId: _taxpayerId, incomeId: _incomeId);
    final restored = await repository.restore(
      taxpayerId: _taxpayerId,
      incomeId: _incomeId,
    );

    expect(adapter.requests.map((r) => r.method), ['DELETE', 'POST']);
    expect(restored.id, _incomeId);
  });

  test('a spent trial cap arrives as a 402 naming incomes_remaining', () async {
    adapter.reply(_incomes, (
      status: 402,
      body: {
        'error': {
          'code': 'entitlement_required',
          'message': 'Trial cap reached.',
          'request_id': 'req_1',
          'details': [
            {
              'field': 'incomes_remaining',
              'code': 'trial_cap_reached',
              'message': 'Trial cap reached.',
            },
          ],
        },
      },
    ));

    final failure = await failureOf(
      () => repository.create(
        taxpayerId: _taxpayerId,
        income: const IncomeCreate(
          regimeCode: '612',
          counterpartyType: CounterpartyType.personaFisica,
          amountCents: 100,
          ivaRate: IvaRate.exempt,
          ivaCents: 0,
        ),
      ),
    );

    expect(
      (failure! as ApiEntitlementRequired).trialCapOf,
      'incomes_remaining',
    );
  });
}
