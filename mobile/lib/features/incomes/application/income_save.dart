import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/api/generated/models/income_create.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/incomes/application/income_records.dart';

/// How a save of the income form (new or edit) ended, in the terms the
/// page acts on.
sealed class IncomeSaveOutcome {
  const new();
}

/// Stored (`201`, or `200` for a retry of the same id).
final class IncomeSaved extends IncomeSaveOutcome {
  /// The stored income, with its non-blocking warnings.
  const new(this.income);

  /// The income as the server answered it.
  final Income income;
}

/// `402`: a plan is needed. [isTrialCap] when the trial's 5 incomes are
/// spent (the account is not read-only).
final class IncomeNeedsPlan extends IncomeSaveOutcome {
  /// The refusal.
  const new({required this.isTrialCap});

  /// Whether the incomes trial cap, not a read-only account, refused it.
  final bool isTrialCap;
}

/// `400`/`422`: the server's verdict on fields ([details]).
final class IncomeRejected extends IncomeSaveOutcome {
  /// The verdict.
  const new(this.details);

  /// One entry per rejected field.
  final List<api.FieldError> details;
}

/// A form-level reason the save did not go through.
enum IncomeSaveNotice {
  /// Offline or timed out: the retry resends the same id.
  unreachable,

  /// `409 id_reused`: this session's id is already an income with other
  /// data (the first answer was lost and the form changed since).
  idReused,

  /// `404` on an edit: the income was deleted (or purged) elsewhere.
  gone,

  /// `429` or anything else.
  failed,
}

/// See [IncomeSaveNotice].
final class IncomeNotSaved extends IncomeSaveOutcome {
  /// The reason.
  const new(this.notice);

  /// Why.
  final IncomeSaveNotice notice;
}

/// The trial cap of incomes in a `402` (`details[].field`).
const String incomesTrialCap = 'incomes_remaining';

/// Saves [income] for [taxpayerId] through [records]; never throws an
/// `ApiFailure`.
Future<IncomeSaveOutcome> saveIncome(
  IncomeRecords records,
  String taxpayerId,
  IncomeCreate income,
) => _outcomeOf(
  () => records.create(taxpayerId, income),
  notFound: IncomeSaveNotice.failed,
);

/// Edits [incomeId] with the merge-patch [patch] through [records]; never
/// throws an `ApiFailure`.
Future<IncomeSaveOutcome> updateIncome(
  IncomeRecords records,
  String taxpayerId,
  String incomeId,
  Map<String, Object?> patch,
) => _outcomeOf(
  () => records.update(taxpayerId, incomeId, patch),
  notFound: IncomeSaveNotice.gone,
);

/// Runs [write]; a `404` becomes [notFound] (the taxpayer is gone on a
/// create, the income on an edit).
Future<IncomeSaveOutcome> _outcomeOf(
  Future<Income> Function() write, {
  required IncomeSaveNotice notFound,
}) async {
  try {
    return IncomeSaved(await write());
  } on ApiEntitlementRequired catch (failure) {
    return IncomeNeedsPlan(isTrialCap: failure.trialCapOf == incomesTrialCap);
  } on ApiValidationFailure catch (failure) {
    return IncomeRejected(failure.details);
  } on ApiRejected catch (failure) {
    if (failure.code == ErrorCode.conflict) {
      return const IncomeNotSaved(IncomeSaveNotice.idReused);
    }
    return IncomeRejected(failure.details);
  } on ApiNotFound {
    return IncomeNotSaved(notFound);
  } on ApiUnreachable {
    return const IncomeNotSaved(IncomeSaveNotice.unreachable);
  } on ApiFailure {
    return const IncomeNotSaved(IncomeSaveNotice.failed);
  }
}
