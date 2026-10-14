import 'package:control_contable/api/generated/models/regime_estimate_status.dart';
import 'package:control_contable/api/generated/models/taxpayer_estimate.dart';

/// The app's own estimate of the provisional ISR payment of [regime] in the
/// month of [estimate] (MODULE-tax-engine §8, OQ-T1): a suggestion for a
/// prior payment the user has not entered, never assumed.
///
/// `null` when there is nothing trustworthy to offer: the regime is not in
/// that month's estimate, has no ISR, or its estimate is itself incomplete
/// (an earlier payment still missing counts as zero there, so the figure
/// would overstate what was due) or only informational.
int? suggestedPriorPayment(TaxpayerEstimate estimate, String regime) {
  final match = estimate.regimes
      .where((candidate) => candidate.regimeCode == regime)
      .firstOrNull;
  final isr = match?.isr;
  if (match == null || isr == null) return null;
  return switch (match.status) {
    RegimeEstimateStatus.complete ||
    RegimeEstimateStatus.partial => isr.payableRoundedCents,
    RegimeEstimateStatus.incomplete ||
    RegimeEstimateStatus.informational ||
    RegimeEstimateStatus.$unknown => null,
  };
}
