import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/api/generated/models/taxpayer_regime.dart';

/// The regimes [taxpayer] holds at some point of [period] (`YYYY-MM`).
List<TaxpayerRegime> regimesHeldIn(Taxpayer taxpayer, String period) => [
  for (final regime in taxpayer.regimes)
    if (regime.startPeriod.compareTo(period) <= 0 &&
        (regime.endPeriod == null || regime.endPeriod!.compareTo(period) >= 0))
      regime,
];
