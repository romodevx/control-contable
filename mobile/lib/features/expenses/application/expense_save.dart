import 'package:control_contable/api/generated/models/expense.dart';
import 'package:control_contable/api/generated/models/expense_create.dart';
import 'package:control_contable/core/records/record_save.dart';
import 'package:control_contable/features/expenses/application/expense_records.dart';

/// How a save of the expense form ended.
typedef ExpenseSaveOutcome = RecordSaveOutcome<Expense>;

/// The trial cap of expenses in a `402` (`details[].field`).
const String expensesTrialCap = 'expenses_remaining';

/// Saves [expense] for [taxpayerId] through [records]; never throws an
/// `ApiFailure`.
Future<ExpenseSaveOutcome> saveExpense(
  ExpenseRecords records,
  String taxpayerId,
  ExpenseCreate expense,
) => recordSaveOutcomeOf(
  () => records.create(taxpayerId, expense),
  notFound: RecordSaveNotice.failed,
  trialCap: expensesTrialCap,
);

/// Edits [expenseId] with the merge-patch [patch] through [records];
/// never throws an `ApiFailure`.
Future<ExpenseSaveOutcome> updateExpense(
  ExpenseRecords records,
  String taxpayerId,
  String expenseId,
  Map<String, Object?> patch,
) => recordSaveOutcomeOf(
  () => records.update(taxpayerId, expenseId, patch),
  notFound: RecordSaveNotice.gone,
  trialCap: expensesTrialCap,
);
