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
