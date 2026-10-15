import 'package:control_contable/api/generated/models/history_viewed_parameter_set.dart';
import 'package:control_contable/api/generated/models/history_viewed_request.dart';
import 'package:control_contable/api/generated/models/taxpayer_estimate.dart';

/// What to post as "seen" for a rendered [estimate], copied from it
/// verbatim (MODULE-history §2.5), or `null` when nothing is to be marked:
/// the [current] month (`YYYY-MM`) or a later one is live and never marked
/// (OQ-H3). Every estimate carries `figures_fingerprint` (required since
/// M8-03, contract 0.11.1).
HistoryViewedRequest? viewedRequestOf(
  TaxpayerEstimate estimate, {
  required String current,
}) {
  if (!_isPast(estimate.period, current)) return null;
  return HistoryViewedRequest(
    recordsFingerprint: estimate.recordsFingerprint,
    figuresFingerprint: estimate.figuresFingerprint,
    engineVersion: estimate.engineVersion,
    parameterSet: HistoryViewedParameterSet(
      fiscalYear: estimate.parameterSet.fiscalYear,
      version: estimate.parameterSet.version,
    ),
  );
}

/// `YYYY-MM` periods order as text.
bool _isPast(String period, String current) => period.compareTo(current) < 0;
