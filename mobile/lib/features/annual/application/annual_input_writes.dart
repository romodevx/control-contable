import 'package:control_contable/api/generated/models/annual_inputs.dart';
import 'package:control_contable/api/generated/models/annual_inputs_update.dart';
import 'package:control_contable/api/generated/models/filed_payments.dart';
import 'package:control_contable/api/generated/models/filed_payments_update.dart';
import 'package:control_contable/api/generated/models/personal_deduction.dart';
import 'package:control_contable/api/generated/models/personal_deduction_create.dart';
import 'package:control_contable/api/generated/models/salary_statement.dart';
import 'package:control_contable/api/generated/models/salary_statement_create.dart';
import 'package:control_contable/core/access/read_only_access.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/records/stacked_undo.dart';
import 'package:control_contable/core/session/session_scoped.dart';
import 'package:control_contable/features/annual/application/annual_estimate.dart';
import 'package:control_contable/features/annual/application/annual_input_data.dart';
import 'package:control_contable/features/annual/data/annual_inputs_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';

/// The writes of the annual inputs. A `402` marks the session read-only
/// (annual writes never use a trial slot, OQ-A9); after every write the
/// list it touched and the estimate that reads it reload. Each method
/// throws the `ApiFailure`.
class AnnualInputWrites {
  /// The writes over [_ref].
  new(this._ref);

  final Ref _ref;

  AnnualInputsRepository get _repository =>
      _ref.read(annualInputsRepositoryProvider);

  /// Records [deduction] for [taxpayerId].
  Future<PersonalDeduction> createDeduction(
    String taxpayerId,
    PersonalDeductionCreate deduction,
  ) => _write(
    () => _repository.createDeduction(
      taxpayerId: taxpayerId,
      deduction: deduction,
    ),
    personalDeductionsProvider,
  );

  /// Deletes [deduction] (allowed while read-only).
  Future<void> deleteDeduction(PersonalDeduction deduction) => _write(
    () => _repository.deleteDeduction(
      taxpayerId: deduction.taxpayerId,
      deductionId: deduction.id,
    ),
    personalDeductionsProvider,
  );

  /// Restores every one of [deductions], going on past a refusal.
  Future<RestoreReport<ApiFailure>> restoreDeductions(
    List<PersonalDeduction> deductions,
  ) => _write(
    () => restoreEach<PersonalDeduction, ApiFailure>(
      deductions,
      (deduction) => _repository.restoreDeduction(
        taxpayerId: deduction.taxpayerId,
        deductionId: deduction.id,
      ),
    ),
    personalDeductionsProvider,
  );

  /// Records [statement] for [taxpayerId].
  Future<SalaryStatement> createStatement(
    String taxpayerId,
    SalaryStatementCreate statement,
  ) => _write(
    () => _repository.createStatement(
      taxpayerId: taxpayerId,
      statement: statement,
    ),
    salaryStatementsProvider,
  );

  /// Edits [statementId] with the merge-patch [patch].
  Future<SalaryStatement> updateStatement(
    String taxpayerId,
    String statementId,
    Map<String, Object?> patch,
  ) => _write(
    () => _repository.updateStatement(
      taxpayerId: taxpayerId,
      statementId: statementId,
      patch: patch,
    ),
    salaryStatementsProvider,
  );

  /// Deletes [statement] for good.
  Future<void> deleteStatement(SalaryStatement statement) => _write(
    () => _repository.deleteStatement(
      taxpayerId: statement.taxpayerId,
      statementId: statement.id,
    ),
    salaryStatementsProvider,
  );

  /// Replaces the annual facts of [query].
  Future<AnnualInputs> putFacts(AnnualQuery query, AnnualInputsUpdate facts) =>
      _write(
        () => _repository.putFacts(
          taxpayerId: query.taxpayerId,
          fiscalYear: query.fiscalYear,
          facts: facts,
        ),
        annualFactsProvider,
      );

  /// Replaces the confirmed months of [query].
  Future<FiledPayments> putFiledPayments(
    FiledPaymentsQuery query,
    FiledPaymentsUpdate payments,
  ) => _write(
    () => _repository.putFiledPayments(
      taxpayerId: query.taxpayerId,
      fiscalYear: query.fiscalYear,
      regime: query.regime,
      payments: payments,
    ),
    filedPaymentsProvider,
  );

  Future<T> _write<T>(
    Future<T> Function() call,
    ProviderOrFamily touched,
  ) async {
    try {
      return await call();
    } on ApiEntitlementRequired {
      _ref.read(readOnlyAccessProvider.notifier).markReadOnly();
      rethrow;
    } finally {
      _ref
        ..invalidate(touched)
        ..invalidate(annualEstimateProvider)
        ..invalidate(annualBreakdownProvider)
        ..invalidate(annualInputCountsProvider);
    }
  }
}

/// See [AnnualInputWrites].
final annualInputWritesProvider = Provider<AnnualInputWrites>(
  AnnualInputWrites.new,
  name: 'annualInputWritesProvider',
);

/// The deductions the Undo toast would restore (it outlives the list).
final annualDeductionsUndoProvider = Provider<StackedUndo<PersonalDeduction>>((
  ref,
) {
  forgetOnSignOut(ref);
  return StackedUndo<PersonalDeduction>();
}, name: 'annualDeductionsUndoProvider');
