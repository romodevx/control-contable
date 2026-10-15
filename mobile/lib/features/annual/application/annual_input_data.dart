import 'package:control_contable/api/generated/models/annual_inputs.dart';
import 'package:control_contable/api/generated/models/filed_payments.dart';
import 'package:control_contable/api/generated/models/personal_deduction_list.dart';
import 'package:control_contable/api/generated/models/salary_statement_list.dart';
import 'package:control_contable/features/annual/application/annual_estimate.dart';
import 'package:control_contable/features/annual/data/annual_inputs_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';

/// Which regime's filed payments to read.
typedef FiledPaymentsQuery = ({
  String taxpayerId,
  int fiscalYear,
  String regime,
});

/// The personal deductions of one taxpayer and year (ANN-19).
final FutureProviderFamily<PersonalDeductionList, AnnualQuery>
personalDeductionsProvider = FutureProvider.autoDispose
    .family<PersonalDeductionList, AnnualQuery>(
      (ref, query) => ref
          .read(annualInputsRepositoryProvider)
          .listDeductions(
            taxpayerId: query.taxpayerId,
            fiscalYear: query.fiscalYear,
          ),
      name: 'personalDeductionsProvider',
      retry: _noAutomaticRetry,
    );

/// The salary constancias of one taxpayer and year (ANN-20).
final FutureProviderFamily<SalaryStatementList, AnnualQuery>
salaryStatementsProvider = FutureProvider.autoDispose
    .family<SalaryStatementList, AnnualQuery>(
      (ref, query) => ref
          .read(annualInputsRepositoryProvider)
          .listStatements(
            taxpayerId: query.taxpayerId,
            fiscalYear: query.fiscalYear,
          ),
      name: 'salaryStatementsProvider',
      retry: _noAutomaticRetry,
    );

/// The per-year annual facts of one taxpayer (ANN-21).
final FutureProviderFamily<AnnualInputs, AnnualQuery> annualFactsProvider =
    FutureProvider.autoDispose.family<AnnualInputs, AnnualQuery>(
      (ref, query) => ref
          .read(annualInputsRepositoryProvider)
          .getFacts(taxpayerId: query.taxpayerId, fiscalYear: query.fiscalYear),
      name: 'annualFactsProvider',
      retry: _noAutomaticRetry,
    );

/// The twelve months of provisional payments of one regime (ANN-21).
final FutureProviderFamily<FiledPayments, FiledPaymentsQuery>
filedPaymentsProvider = FutureProvider.autoDispose
    .family<FiledPayments, FiledPaymentsQuery>(
      (ref, query) => ref
          .read(annualInputsRepositoryProvider)
          .getFiledPayments(
            taxpayerId: query.taxpayerId,
            fiscalYear: query.fiscalYear,
            regime: query.regime,
          ),
      name: 'filedPaymentsProvider',
      retry: _noAutomaticRetry,
    );

/// A failure shows a retry: no automatic retry (Riverpod 3 retries by
/// default).
Duration? _noAutomaticRetry(int retryCount, Object error) => null;
