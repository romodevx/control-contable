import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/core/network/api_failure.dart';

/// How a save of a record form (new or edit) ended, in the terms the page
/// acts on.
sealed class RecordSaveOutcome<T> {
  const new();
}

/// Stored (`201`, or `200` for a retry of the same id or an edit).
final class RecordSaved<T> extends RecordSaveOutcome<T> {
  /// The stored record, with its non-blocking warnings.
  const new(this.record);

  /// The record as the server answered it.
  final T record;
}

/// `402`: a plan is needed. [isTrialCap] when the trial's records of this
/// kind are spent (the account is not read-only).
final class RecordNeedsPlan<T> extends RecordSaveOutcome<T> {
  /// The refusal.
  const new({required this.isTrialCap});

  /// Whether the module's trial cap, not a read-only account, refused it.
  final bool isTrialCap;
}

/// `400`/`422`: the server's verdict on fields ([details]).
final class RecordRejected<T> extends RecordSaveOutcome<T> {
  /// The verdict.
  const new(this.details);

  /// One entry per rejected field.
  final List<api.FieldError> details;
}

/// A form-level reason the save did not go through.
enum RecordSaveNotice {
  /// Offline or timed out: the retry resends the same id.
  unreachable,

  /// `409 id_reused`: this session's id is already a record with other
  /// data (the first answer was lost and the form changed since).
  idReused,

  /// `404` on an edit: the record was deleted (or purged) elsewhere.
  gone,

  /// `429` or anything else.
  failed,
}

/// See [RecordSaveNotice].
final class RecordNotSaved<T> extends RecordSaveOutcome<T> {
  /// The reason.
  const new(this.notice);

  /// Why.
  final RecordSaveNotice notice;
}

/// Runs [write] and names how it ended; never throws an `ApiFailure`. A
/// `404` becomes [notFound] (the taxpayer is gone on a create, the record
/// on an edit); a `402` naming [trialCap] is the module's trial cap.
Future<RecordSaveOutcome<T>> recordSaveOutcomeOf<T>(
  Future<T> Function() write, {
  required RecordSaveNotice notFound,
  required String trialCap,
}) async {
  try {
    return RecordSaved(await write());
  } on ApiEntitlementRequired catch (failure) {
    return RecordNeedsPlan(isTrialCap: failure.trialCapOf == trialCap);
  } on ApiValidationFailure catch (failure) {
    return RecordRejected(failure.details);
  } on ApiRejected catch (failure) {
    if (failure.code == ErrorCode.conflict) {
      return const RecordNotSaved(RecordSaveNotice.idReused);
    }
    return RecordRejected(failure.details);
  } on ApiNotFound {
    return RecordNotSaved(notFound);
  } on ApiUnreachable {
    return const RecordNotSaved(RecordSaveNotice.unreachable);
  } on ApiFailure {
    return const RecordNotSaved(RecordSaveNotice.failed);
  }
}
