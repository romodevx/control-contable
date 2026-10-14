import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/api/generated/models/tax_input_value.dart';
import 'package:control_contable/api/generated/models/tax_inputs.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/tax_engine/data/tax_engine_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';

/// Which tax inputs to read: a taxpayer, a regime and a fiscal year.
typedef TaxInputsQuery = ({String taxpayerId, String regime, int fiscalYear});

/// The stored tax inputs of one regime and fiscal year, with their
/// definitions. A failure shows a retry: no automatic retry.
final FutureProviderFamily<TaxInputs, TaxInputsQuery> taxInputsProvider =
    FutureProvider.autoDispose.family<TaxInputs, TaxInputsQuery>(
      (ref, query) => ref
          .read(taxEngineRepositoryProvider)
          .getTaxInputs(
            taxpayerId: query.taxpayerId,
            regime: query.regime,
            fiscalYear: query.fiscalYear,
          ),
      name: 'taxInputsProvider',
      retry: (retryCount, error) => null,
    );

/// How a save of the tax-inputs form ended, in the terms the page acts on.
sealed class TaxInputsSaveOutcome {
  const new();
}

/// Stored (`200`).
final class TaxInputsSaved extends TaxInputsSaveOutcome {
  /// The stored inputs.
  const new(this.inputs);

  /// What the server stored.
  final TaxInputs inputs;
}

/// `402`: the account is read-only.
final class TaxInputsNeedPlan extends TaxInputsSaveOutcome {
  /// The refusal.
  const new();
}

/// `400`: the server's verdict on values (`details[].field` is
/// `values.<index>[.<property>]`).
final class TaxInputsRejected extends TaxInputsSaveOutcome {
  /// The verdict.
  const new(this.details);

  /// One entry per rejected value.
  final List<api.FieldError> details;
}

/// A form-level reason the save did not go through.
enum TaxInputsSaveNotice {
  /// Offline or timed out: saving again is safe (the PUT is idempotent).
  unreachable,

  /// `422 taxpayer_archived`.
  archived,

  /// `422 regime_not_active`: the regime is not held that fiscal year.
  regimeNotActive,

  /// `422 parameters_unavailable`: a fiscal year before 2026.
  parametersUnavailable,

  /// `404`: the taxpayer is gone.
  gone,

  /// `429` or anything else.
  failed,
}

/// See [TaxInputsSaveNotice].
final class TaxInputsNotSaved extends TaxInputsSaveOutcome {
  /// The reason.
  const new(this.notice);

  /// Why.
  final TaxInputsSaveNotice notice;
}

const Map<String, TaxInputsSaveNotice> _ruleNotices = {
  'taxpayer_archived': TaxInputsSaveNotice.archived,
  'regime_not_active': TaxInputsSaveNotice.regimeNotActive,
  'parameters_unavailable': TaxInputsSaveNotice.parametersUnavailable,
};

/// Replaces the inputs of [query] with [values] through [repository];
/// never throws an `ApiFailure`.
Future<TaxInputsSaveOutcome> saveTaxInputs(
  TaxEngineRepository repository,
  TaxInputsQuery query,
  List<TaxInputValue> values,
) async {
  try {
    return TaxInputsSaved(
      await repository.putTaxInputs(
        taxpayerId: query.taxpayerId,
        regime: query.regime,
        fiscalYear: query.fiscalYear,
        values: values,
      ),
    );
  } on ApiEntitlementRequired {
    return const TaxInputsNeedPlan();
  } on ApiValidationFailure catch (failure) {
    return TaxInputsRejected(failure.details);
  } on ApiRejected catch (failure) {
    final notice = failure.details
        .map((detail) => _ruleNotices[detail.code])
        .nonNulls
        .firstOrNull;
    return TaxInputsNotSaved(notice ?? TaxInputsSaveNotice.failed);
  } on ApiNotFound {
    return const TaxInputsNotSaved(TaxInputsSaveNotice.gone);
  } on ApiUnreachable {
    return const TaxInputsNotSaved(TaxInputsSaveNotice.unreachable);
  } on ApiFailure {
    return const TaxInputsNotSaved(TaxInputsSaveNotice.failed);
  }
}
