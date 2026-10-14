import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/api/generated/models/income_create.dart';
import 'package:control_contable/core/records/record_save.dart';
import 'package:control_contable/features/incomes/application/income_records.dart';

/// How a save of the income form ended.
typedef IncomeSaveOutcome = RecordSaveOutcome<Income>;

/// The trial cap of incomes in a `402` (`details[].field`).
const String incomesTrialCap = 'incomes_remaining';

/// Saves [income] for [taxpayerId] through [records]; never throws an
/// `ApiFailure`.
Future<IncomeSaveOutcome> saveIncome(
  IncomeRecords records,
  String taxpayerId,
  IncomeCreate income,
) => recordSaveOutcomeOf(
  () => records.create(taxpayerId, income),
  notFound: RecordSaveNotice.failed,
  trialCap: incomesTrialCap,
);

/// Edits [incomeId] with the merge-patch [patch] through [records]; never
/// throws an `ApiFailure`.
Future<IncomeSaveOutcome> updateIncome(
  IncomeRecords records,
  String taxpayerId,
  String incomeId,
  Map<String, Object?> patch,
) => recordSaveOutcomeOf(
  () => records.update(taxpayerId, incomeId, patch),
  notFound: RecordSaveNotice.gone,
  trialCap: incomesTrialCap,
);
