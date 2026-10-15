import 'package:control_contable/api/generated/api_client.dart';
import 'package:control_contable/api/generated/models/annual_inputs_update.dart';
import 'package:control_contable/api/generated/models/deduction_beneficiary.dart';
import 'package:control_contable/api/generated/models/deduction_category.dart';
import 'package:control_contable/api/generated/models/deduction_payment_method.dart';
import 'package:control_contable/api/generated/models/filed_payment_entry.dart';
import 'package:control_contable/api/generated/models/filed_payments_update.dart';
import 'package:control_contable/api/generated/models/personal_deduction_create.dart';
import 'package:control_contable/api/generated/models/salary_statement_create.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/annual/data/annual_inputs_repository.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/recording_adapter.dart';
import '../tax_engine/estimate_fixtures.dart' show fixtureTaxpayerId;
import 'annual_inputs_fixtures.dart';

const String _base = '/v1/taxpayers/$fixtureTaxpayerId';
const String _deductions = '$_base/personal-deductions';
const String _statements = '$_base/salary-statements';

void main() {
  late RecordingAdapter adapter;
  late ApiAnnualInputsRepository repository;

  setUp(() {
    adapter = RecordingAdapter(
      replies: {
        _deductions: (
          status: 200,
          body: {
            'fiscal_year': 2026,
            'items': [deductionJson()],
            'totals': <Object>[],
          },
        ),
        '$_deductions/ded-1': (status: 204, body: null),
        '$_deductions/ded-1/restore': (status: 200, body: deductionJson()),
        _statements: (
          status: 200,
          body: {
            'fiscal_year': 2026,
            'items': [statementJson()],
            'totals': {
              'accumulable_cents': 1,
              'exempt_cents': 0,
              'isr_withheld_cents': 0,
            },
          },
        ),
        '$_statements/sal-1': (status: 200, body: statementJson()),
        '$_base/annual-inputs/2026': (
          status: 200,
          body: annualFactsOf().toJson(),
        ),
        '$_base/regimes/612/filed-payments/2026': (
          status: 200,
          body: filedPaymentsOf().toJson(),
        ),
      },
    );
    final dio = Dio(BaseOptions(baseUrl: 'http://localhost:3001'))
      ..httpClientAdapter = adapter;
    repository = ApiAnnualInputsRepository(ApiClient(dio).annual, dio);
  });

  test('lists the deductions of one fiscal year', () async {
    final list = await repository.listDeductions(
      taxpayerId: fixtureTaxpayerId,
      fiscalYear: 2026,
    );

    final request = adapter.requests.single;
    expect(request.uri.path, _deductions);
    expect(request.uri.queryParameters, {'fiscal_year': '2026'});
    expect(list.items.single.id, 'ded-1');
  });

  test('creates, deletes and restores a deduction', () async {
    adapter.reply(_deductions, (status: 201, body: deductionJson()));
    await repository.createDeduction(
      taxpayerId: fixtureTaxpayerId,
      deduction: const PersonalDeductionCreate(
        id: 'ded-1',
        category: DeductionCategory.medical,
        amountCents: 150000,
        paidOn: '2026-03-10',
        paymentMethod: DeductionPaymentMethod.electronic,
        beneficiary: DeductionBeneficiary.self,
      ),
    );
    await repository.deleteDeduction(
      taxpayerId: fixtureTaxpayerId,
      deductionId: 'ded-1',
    );
    await repository.restoreDeduction(
      taxpayerId: fixtureTaxpayerId,
      deductionId: 'ded-1',
    );

    expect(
      [for (final r in adapter.requests) '${r.method} ${r.uri.path}'],
      [
        'POST $_deductions',
        'DELETE $_deductions/ded-1',
        'POST $_deductions/ded-1/restore',
      ],
    );
    expect(adapter.requests.first.data, containsPair('id', 'ded-1'));
  });

  test('salary constancias: list, create, merge-patch with a cleared label, '
      'delete', () async {
    await repository.listStatements(
      taxpayerId: fixtureTaxpayerId,
      fiscalYear: 2026,
    );
    adapter.reply(_statements, (status: 201, body: statementJson()));
    await repository.createStatement(
      taxpayerId: fixtureTaxpayerId,
      statement: const SalaryStatementCreate(
        fiscalYear: 2026,
        accumulableCents: 1,
        exemptCents: 0,
        isrWithheldCents: 0,
      ),
    );
    await repository.updateStatement(
      taxpayerId: fixtureTaxpayerId,
      statementId: 'sal-1',
      patch: const {'employer_label': null},
    );
    adapter.reply('$_statements/sal-1', (status: 204, body: null));
    await repository.deleteStatement(
      taxpayerId: fixtureTaxpayerId,
      statementId: 'sal-1',
    );

    expect(
      [for (final r in adapter.requests) r.method],
      ['GET', 'POST', 'PATCH', 'DELETE'],
    );
    expect(adapter.requests[2].data, {'employer_label': null});
  });

  test('reads and replaces the annual facts', () async {
    final facts = await repository.getFacts(
      taxpayerId: fixtureTaxpayerId,
      fiscalYear: 2026,
    );
    await repository.putFacts(
      taxpayerId: fixtureTaxpayerId,
      fiscalYear: 2026,
      facts: const AnnualInputsUpdate(
        priorYearAccumulableIncomeCents: null,
        otherExemptIncomeCents: 0,
        toldEmployerWillFile: true,
      ),
    );

    expect(facts.regimeTaxInputs, hasLength(3));
    expect(adapter.requests.last.method, 'PUT');
    expect(
      adapter.requests.last.data,
      containsPair('prior_year_accumulable_income_cents', null),
    );
  });

  test('reads and replaces the filed payments of a regime', () async {
    final payments = await repository.getFiledPayments(
      taxpayerId: fixtureTaxpayerId,
      fiscalYear: 2026,
      regime: '612',
    );
    await repository.putFiledPayments(
      taxpayerId: fixtureTaxpayerId,
      fiscalYear: 2026,
      regime: '612',
      payments: const FiledPaymentsUpdate(
        months: [FiledPaymentEntry(month: 1, amountCents: 100000)],
      ),
    );

    expect(payments.months, hasLength(12));
    expect(adapter.requests.last.method, 'PUT');
  });

  test('a refusal is an ApiFailure', () async {
    adapter.reply(_deductions, (
      status: 402,
      body: {
        'error': {
          'code': 'entitlement_required',
          'message': 'Read only.',
          'details': <Object>[],
        },
      },
    ));

    final failure = await failureOf(
      () => repository.listDeductions(
        taxpayerId: fixtureTaxpayerId,
        fiscalYear: 2026,
      ),
    );

    expect(failure, isA<ApiEntitlementRequired>());
  });
}
