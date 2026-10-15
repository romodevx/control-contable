import 'package:control_contable/api/generated/clients/annual_client.dart';
import 'package:control_contable/api/generated/models/annual_inputs.dart';
import 'package:control_contable/api/generated/models/annual_inputs_update.dart';
import 'package:control_contable/api/generated/models/filed_payments.dart';
import 'package:control_contable/api/generated/models/filed_payments_update.dart';
import 'package:control_contable/api/generated/models/personal_deduction.dart';
import 'package:control_contable/api/generated/models/personal_deduction_create.dart';
import 'package:control_contable/api/generated/models/personal_deduction_list.dart';
import 'package:control_contable/api/generated/models/salary_statement.dart';
import 'package:control_contable/api/generated/models/salary_statement_create.dart';
import 'package:control_contable/api/generated/models/salary_statement_list.dart';
import 'package:control_contable/core/network/api_client_provider.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/network/dio_provider.dart';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The annual inputs of a taxpayer and fiscal year (ANN-19..21): personal
/// deductions, salary constancias, filed provisional payments and the
/// per-year facts. Every method throws an `ApiFailure` when the server
/// refuses or cannot be reached.
abstract interface class AnnualInputsRepository {
  /// The live personal deductions of [fiscalYear], with registered totals.
  Future<PersonalDeductionList> listDeductions({
    required String taxpayerId,
    required int fiscalYear,
  });

  /// Records [deduction] (idempotent on its client id).
  Future<PersonalDeduction> createDeduction({
    required String taxpayerId,
    required PersonalDeductionCreate deduction,
  });

  /// Soft-deletes a deduction (allowed while read-only).
  Future<void> deleteDeduction({
    required String taxpayerId,
    required String deductionId,
  });

  /// Undoes a delete.
  Future<PersonalDeduction> restoreDeduction({
    required String taxpayerId,
    required String deductionId,
  });

  /// The salary constancias of [fiscalYear], with totals.
  Future<SalaryStatementList> listStatements({
    required String taxpayerId,
    required int fiscalYear,
  });

  /// Records [statement] (idempotent on its client id).
  Future<SalaryStatement> createStatement({
    required String taxpayerId,
    required SalaryStatementCreate statement,
  });

  /// Edits a constancia with the merge-patch [patch] (`null` clears).
  Future<SalaryStatement> updateStatement({
    required String taxpayerId,
    required String statementId,
    required Map<String, Object?> patch,
  });

  /// Deletes a constancia for good (the screen confirms first).
  Future<void> deleteStatement({
    required String taxpayerId,
    required String statementId,
  });

  /// The per-year facts, plus the regimes' annual tax inputs (read-only).
  Future<AnnualInputs> getFacts({
    required String taxpayerId,
    required int fiscalYear,
  });

  /// Replaces the per-year facts.
  Future<AnnualInputs> putFacts({
    required String taxpayerId,
    required int fiscalYear,
    required AnnualInputsUpdate facts,
  });

  /// The twelve months of provisional payments of [regime].
  Future<FiledPayments> getFiledPayments({
    required String taxpayerId,
    required int fiscalYear,
    required String regime,
  });

  /// Replaces the confirmed months of [regime] (absent = not confirmed).
  Future<FiledPayments> putFiledPayments({
    required String taxpayerId,
    required int fiscalYear,
    required String regime,
    required FiledPaymentsUpdate payments,
  });
}

/// [AnnualInputsRepository] over the generated client, except
/// [updateStatement]: the generated update cannot tell "unset" from
/// "clear", so the merge-patch travels as a map (ADR-0003 "Null on the
/// wire").
class ApiAnnualInputsRepository implements AnnualInputsRepository {
  /// The repository on [_client] and its transport [_dio].
  new(this._client, this._dio);

  final AnnualClient _client;
  final Dio _dio;

  @override
  Future<PersonalDeductionList> listDeductions({
    required String taxpayerId,
    required int fiscalYear,
  }) => guardApi(
    () => _client.listPersonalDeductions(
      taxpayerId: taxpayerId,
      fiscalYear: fiscalYear,
    ),
  );

  @override
  Future<PersonalDeduction> createDeduction({
    required String taxpayerId,
    required PersonalDeductionCreate deduction,
  }) => guardApi(
    () => _client.createPersonalDeduction(
      taxpayerId: taxpayerId,
      body: deduction,
    ),
  );

  @override
  Future<void> deleteDeduction({
    required String taxpayerId,
    required String deductionId,
  }) => guardApi(
    () => _client.deletePersonalDeduction(
      taxpayerId: taxpayerId,
      deductionId: deductionId,
    ),
  );

  @override
  Future<PersonalDeduction> restoreDeduction({
    required String taxpayerId,
    required String deductionId,
  }) => guardApi(
    () => _client.restorePersonalDeduction(
      taxpayerId: taxpayerId,
      deductionId: deductionId,
    ),
  );

  @override
  Future<SalaryStatementList> listStatements({
    required String taxpayerId,
    required int fiscalYear,
  }) => guardApi(
    () => _client.listSalaryStatements(
      taxpayerId: taxpayerId,
      fiscalYear: fiscalYear,
    ),
  );

  @override
  Future<SalaryStatement> createStatement({
    required String taxpayerId,
    required SalaryStatementCreate statement,
  }) => guardApi(
    () =>
        _client.createSalaryStatement(taxpayerId: taxpayerId, body: statement),
  );

  @override
  Future<SalaryStatement> updateStatement({
    required String taxpayerId,
    required String statementId,
    required Map<String, Object?> patch,
  }) => guardApi(() async {
    final response = await _dio.patch<Map<String, Object?>>(
      '/v1/taxpayers/${Uri.encodeComponent(taxpayerId)}'
      '/salary-statements/${Uri.encodeComponent(statementId)}',
      data: patch,
    );
    return SalaryStatement.fromJson(response.data!);
  });

  @override
  Future<void> deleteStatement({
    required String taxpayerId,
    required String statementId,
  }) => guardApi(
    () => _client.deleteSalaryStatement(
      taxpayerId: taxpayerId,
      statementId: statementId,
    ),
  );

  @override
  Future<AnnualInputs> getFacts({
    required String taxpayerId,
    required int fiscalYear,
  }) => guardApi(
    () =>
        _client.getAnnualInputs(taxpayerId: taxpayerId, fiscalYear: fiscalYear),
  );

  @override
  Future<AnnualInputs> putFacts({
    required String taxpayerId,
    required int fiscalYear,
    required AnnualInputsUpdate facts,
  }) => guardApi(
    () => _client.putAnnualInputs(
      taxpayerId: taxpayerId,
      fiscalYear: fiscalYear,
      body: facts,
    ),
  );

  @override
  Future<FiledPayments> getFiledPayments({
    required String taxpayerId,
    required int fiscalYear,
    required String regime,
  }) => guardApi(
    () => _client.getFiledPayments(
      taxpayerId: taxpayerId,
      regimeCode: regime,
      fiscalYear: fiscalYear,
    ),
  );

  @override
  Future<FiledPayments> putFiledPayments({
    required String taxpayerId,
    required int fiscalYear,
    required String regime,
    required FiledPaymentsUpdate payments,
  }) => guardApi(
    () => _client.putFiledPayments(
      taxpayerId: taxpayerId,
      regimeCode: regime,
      fiscalYear: fiscalYear,
      body: payments,
    ),
  );
}

/// The annual inputs operations. Tests override them with a fake.
final annualInputsRepositoryProvider = Provider<AnnualInputsRepository>(
  (ref) => ApiAnnualInputsRepository(
    ref.watch(apiClientProvider).annual,
    ref.watch(dioProvider),
  ),
  name: 'annualInputsRepositoryProvider',
);
