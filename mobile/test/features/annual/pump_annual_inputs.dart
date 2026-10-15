import 'dart:async';

import 'package:control_contable/api/generated/models/annual_inputs.dart';
import 'package:control_contable/api/generated/models/annual_inputs_update.dart';
import 'package:control_contable/api/generated/models/filed_payments.dart';
import 'package:control_contable/api/generated/models/filed_payments_update.dart';
import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/personal_deduction.dart';
import 'package:control_contable/api/generated/models/personal_deduction_create.dart';
import 'package:control_contable/api/generated/models/personal_deduction_list.dart';
import 'package:control_contable/api/generated/models/salary_statement.dart';
import 'package:control_contable/api/generated/models/salary_statement_create.dart';
import 'package:control_contable/api/generated/models/salary_statement_list.dart';
import 'package:control_contable/core/access/read_only_access.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/annual/data/annual_inputs_repository.dart';
import 'package:control_contable/features/annual/data/annual_repository.dart';
import 'package:control_contable/features/taxpayers/application/client_id.dart';
import 'package:control_contable/features/taxpayers/data/active_taxpayer_store.dart';
import 'package:control_contable/features/taxpayers/data/taxpayers_repository.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/pump_app.dart';
import '../taxpayers/taxpayers_fakes.dart';
import 'annual_inputs_fixtures.dart';
import 'pump_annual.dart';

/// An in-memory annual inputs server. Each operation named in [failures]
/// fails with that exception (or waits on that [Completer]) until removed;
/// every write is recorded.
class FakeAnnualInputsRepository implements AnnualInputsRepository {
  /// The live deductions, as JSON.
  final List<Map<String, Object?>> deductions = [];

  /// The constancias, as JSON.
  final List<Map<String, Object?>> statements = [];

  /// The annual facts.
  AnnualInputs facts = annualFactsOf();

  /// The filed payments per regime.
  final Map<String, FiledPayments> payments = {'612': filedPaymentsOf()};

  /// Operation name → what it answers instead (an exception or a
  /// `Completer` still pending).
  final Map<String, Object> failures = {};

  /// Every create body sent.
  final List<PersonalDeductionCreate> createdDeductions = [];

  /// Every constancia create body sent.
  final List<SalaryStatementCreate> createdStatements = [];

  /// Every constancia patch sent.
  final List<Map<String, Object?>> statementPatches = [];

  /// Every facts body sent.
  final List<AnnualInputsUpdate> putFactsBodies = [];

  /// Every filed payments body sent, by regime.
  final List<(String, FiledPaymentsUpdate)> putPaymentBodies = [];

  /// Every operation, in order (`deleteDeduction ded-1`).
  final List<String> calls = [];

  final Map<String, Map<String, Object?>> _deleted = {};

  Future<T> _run<T>(String operation, T Function() answer) async {
    calls.add(operation);
    final failure = failures[operation.split(' ').first];
    if (failure is Completer<void>) await failure.future;
    if (failure is Exception) throw failure;
    return answer();
  }

  @override
  Future<PersonalDeductionList> listDeductions({
    required String taxpayerId,
    required int fiscalYear,
  }) => _run('listDeductions', () => deductionListOf(deductions));

  @override
  Future<PersonalDeduction> createDeduction({
    required String taxpayerId,
    required PersonalDeductionCreate deduction,
  }) => _run('createDeduction', () {
    createdDeductions.add(deduction);
    final json = {
      ...deductionJson(id: deduction.id ?? 'new'),
      ...deduction.toJson(),
      'excluded':
          deduction.paymentMethod.json == 'cash' &&
              deduction.category.json == 'medical'
          ? 'cash_not_deductible'
          : null,
    };
    deductions.add(json);
    return deductionFrom(json);
  });

  @override
  Future<void> deleteDeduction({
    required String taxpayerId,
    required String deductionId,
  }) => _run('deleteDeduction $deductionId', () {
    final index = deductions.indexWhere((d) => d['id'] == deductionId);
    _deleted[deductionId] = deductions.removeAt(index);
  });

  @override
  Future<PersonalDeduction> restoreDeduction({
    required String taxpayerId,
    required String deductionId,
  }) => _run('restoreDeduction $deductionId', () {
    final json = _deleted.remove(deductionId)!;
    deductions.add(json);
    return deductionFrom(json);
  });

  @override
  Future<SalaryStatementList> listStatements({
    required String taxpayerId,
    required int fiscalYear,
  }) => _run('listStatements', () => statementListOf(statements));

  @override
  Future<SalaryStatement> createStatement({
    required String taxpayerId,
    required SalaryStatementCreate statement,
  }) => _run('createStatement', () {
    createdStatements.add(statement);
    final json = {
      ...statementJson(id: statement.id ?? 'new', employerLabel: null),
      ...statement.toJson(),
    };
    statements.add(json);
    return statementFrom(json);
  });

  @override
  Future<SalaryStatement> updateStatement({
    required String taxpayerId,
    required String statementId,
    required Map<String, Object?> patch,
  }) => _run('updateStatement $statementId', () {
    statementPatches.add(patch);
    final index = statements.indexWhere((s) => s['id'] == statementId);
    statements[index] = {...statements[index], ...patch};
    return statementFrom(statements[index]);
  });

  @override
  Future<void> deleteStatement({
    required String taxpayerId,
    required String statementId,
  }) => _run(
    'deleteStatement $statementId',
    () => statements.removeWhere((s) => s['id'] == statementId),
  );

  @override
  Future<AnnualInputs> getFacts({
    required String taxpayerId,
    required int fiscalYear,
  }) => _run('getFacts', () => facts);

  @override
  Future<AnnualInputs> putFacts({
    required String taxpayerId,
    required int fiscalYear,
    required AnnualInputsUpdate facts,
  }) => _run('putFacts', () {
    putFactsBodies.add(facts);
    return this.facts = this.facts.copyWith(
      priorYearAccumulableIncomeCents: facts.priorYearAccumulableIncomeCents,
      otherExemptIncomeCents: facts.otherExemptIncomeCents,
      toldEmployerWillFile: facts.toldEmployerWillFile,
      updatedAt: DateTime.utc(2026, 10, 14),
    );
  });

  @override
  Future<FiledPayments> getFiledPayments({
    required String taxpayerId,
    required int fiscalYear,
    required String regime,
  }) => _run('getFiledPayments $regime', () => payments[regime]!);

  @override
  Future<FiledPayments> putFiledPayments({
    required String taxpayerId,
    required int fiscalYear,
    required String regime,
    required FiledPaymentsUpdate payments,
  }) => _run('putFiledPayments $regime', () {
    putPaymentBodies.add((regime, payments));
    final filed = {for (final e in payments.months) e.month: e.amountCents};
    final current = this.payments[regime]!;
    return this.payments[regime] = filedPaymentsOf(
      regime: regime,
      saved: true,
      months: [
        for (final month in current.months)
          paymentMonthJson(
            month.month,
            inWindow: month.inWindow,
            filedCents: filed[month.month],
            taxInputCents: month.taxInputCents,
            appEstimateCents: month.appEstimateCents,
            appEstimateIncomplete: month.appEstimateIncomplete,
          ),
      ],
    );
  });
}

/// What an annual inputs widget test scripts and observes.
class AnnualInputsHarness {
  /// The taxpayers server (one taxpayer holding 612).
  final FakeTaxpayersRepository taxpayers = FakeTaxpayersRepository();

  /// The annual estimate server (AN-01 by default).
  final FakeAnnualRepository annual = FakeAnnualRepository();

  /// The annual inputs server.
  final FakeAnnualInputsRepository inputs = FakeAnnualInputsRepository();

  /// The running app's providers.
  late ProviderContainer container;

  /// "Now": 2026-10-14 noon in Mexico City.
  DateTime now = DateTime.utc(2026, 10, 14, 18);

  /// The active taxpayer's person type.
  PersonType personType = PersonType.fisica;

  /// Whether the session already knows the user is read-only.
  bool readOnly = false;

  int _ids = 0;

  /// Boots the app with `module_annual` [enabled] and opens [location] on
  /// top of home.
  Future<void> pump(
    WidgetTester tester, {
    required String location,
    bool enabled = true,
  }) async {
    taxpayers.taxpayers.add(
      taxpayer(personType: personType, regimes: [heldRegime('612')]),
    );
    await pumpControlContableApp(
      tester,
      enabledModules: {
        AppModule.taxpayers,
        AppModule.taxEngine,
        if (enabled) AppModule.annual,
      },
      overrides: [
        taxpayersRepositoryProvider.overrideWithValue(taxpayers),
        annualRepositoryProvider.overrideWithValue(annual),
        annualInputsRepositoryProvider.overrideWithValue(inputs),
        clockProvider.overrideWithValue(() => now),
        clientIdProvider.overrideWithValue(() => 'client-${++_ids}'),
        activeTaxpayerStoreProvider.overrideWithValue(
          InMemoryActiveTaxpayerStore(),
        ),
      ],
    );
    container = ProviderScope.containerOf(
      tester.element(find.byType(MaterialApp)),
    );
    if (readOnly) {
      container.read(readOnlyAccessProvider.notifier).markReadOnly();
    }
    unawaited(container.read(appRouterProvider).push<void>(location));
    await tester.pumpAndSettle();
  }

  /// The location on top.
  String get location => container.read(appRouterProvider).state.uri.toString();
}
