import 'package:control_contable/api/generated/models/figures_fingerprint.dart';
import 'package:control_contable/api/generated/models/history_viewed_parameter_set.dart';
import 'package:control_contable/api/generated/models/history_viewed_request.dart';
import 'package:control_contable/api/generated/models/taxpayer_estimate.dart';

/// What to post as "seen" for a rendered [estimate], copied from it
/// verbatim (MODULE-history §2.5), or `null` when nothing is to be marked:
/// the [current] month (`YYYY-MM`) or a later one is live and never marked
/// (OQ-H3), and an estimate without a figures fingerprint cannot be
/// compared later (contract 0.9.0, optional until M8-03).
HistoryViewedRequest? viewedRequestOf(
  TaxpayerEstimate estimate, {
  required String current,
}) {
  final figures = _figuresOf(estimate);
  if (figures == null || !_isPast(estimate.period, current)) return null;
  return HistoryViewedRequest(
    recordsFingerprint: estimate.recordsFingerprint,
    figuresFingerprint: figures,
    engineVersion: estimate.engineVersion,
    parameterSet: HistoryViewedParameterSet(
      fiscalYear: estimate.parameterSet.fiscalYear,
      version: estimate.parameterSet.version,
    ),
  );
}

/// Nullable on purpose: the field is optional until M8-03 makes it
/// required, and this keeps compiling (and lint-clean) either way.
FiguresFingerprint? _figuresOf(TaxpayerEstimate estimate) =>
    estimate.figuresFingerprint;

/// `YYYY-MM` periods order as text.
bool _isPast(String period, String current) => period.compareTo(current) < 0;
