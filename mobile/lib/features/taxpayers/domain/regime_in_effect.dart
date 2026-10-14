import 'package:control_contable/api/generated/models/taxpayer_regime.dart';

/// Whether any of [regimes] applies to [period] (`YYYY-MM`): an open one,
/// or one whose `end_period` is [period] or later. Without one there is
/// nothing to estimate for that month.
bool hasRegimeInEffect(
  List<TaxpayerRegime> regimes, {
  required String period,
}) => regimes.any((regime) {
  final end = regime.endPeriod;
  return end == null || end.compareTo(period) >= 0;
});
