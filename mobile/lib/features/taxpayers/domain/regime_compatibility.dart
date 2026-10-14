import 'package:control_contable/api/generated/models/regime.dart';
import 'package:control_contable/api/generated/models/regime_incompatibility.dart';

/// The incompatible pairs among [selected] (in selection order), keyed by
/// the code of the later regime of each pair; the value names the earlier
/// one and the catalog's reason. A pair counts when either side declares it.
///
/// UX only: onboarding regimes all start in the current month, so their
/// windows overlap; the server answers `422 invalid_regime_combination`.
Map<String, RegimeIncompatibility> regimeConflicts(List<Regime> selected) {
  final conflicts = <String, RegimeIncompatibility>{};
  for (final (index, later) in selected.indexed) {
    for (final earlier in selected.take(index)) {
      final reason = _reasonBetween(earlier, later);
      if (reason == null) continue;
      conflicts[later.code] = RegimeIncompatibility(
        code: earlier.code,
        reason: reason,
      );
      break;
    }
  }
  return conflicts;
}

String? _reasonBetween(Regime a, Regime b) {
  for (final (from, to) in [(a, b), (b, a)]) {
    for (final incompatibility in from.incompatibleWith) {
      if (incompatibility.code == to.code) return incompatibility.reason;
    }
  }
  return null;
}

/// The regimes among [others] that [regime] may overlap only with the
/// warning `regime_combination_needs_cpa_review`: a pair the catalog lists
/// in `review_pending_with` on either side (QA M2-12 D3). [regime] itself
/// is never a partner.
List<Regime> reviewPendingPartners(Regime regime, Iterable<Regime> others) => [
  for (final other in others)
    if (other.code != regime.code && _pendingBetween(regime, other)) other,
];

bool _pendingBetween(Regime a, Regime b) => [(a, b), (b, a)].any(
  (pair) => (pair.$1.reviewPendingWith ?? const <RegimeIncompatibility>[]).any(
    (link) => link.code == pair.$2.code,
  ),
);
