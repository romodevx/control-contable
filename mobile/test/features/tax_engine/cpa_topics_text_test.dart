import 'package:control_contable/api/generated/models/cpa_topic.dart';
import 'package:control_contable/app/app.dart';
import 'package:control_contable/features/tax_engine/presentation/regime_estimate_card.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  final l10n = lookupAppLocalizations(appLocale);

  // Contract 0.10.0 appended two withholding topics (ADR-0008). No engine
  // sends them before WH-09, so until WH-11/WH-12 give them their own copy
  // they read like any topic the app does not know yet.
  test('withholding CPA topics read as the generic topic', () {
    expect(
      cpaTopicsText(l10n, [CpaTopic.withholdingRuleUnverified]),
      l10n.cpaTopicOther,
    );
    expect(
      cpaTopicsText(l10n, [CpaTopic.withheldIvaCreditTiming]),
      l10n.cpaTopicOther,
    );
  });

  // Contract 0.11.0 appended the annual topics (ADR-0009). Only the annual
  // estimate sends them; on a monthly card they read as the generic topic
  // until ANN-17 gives the annual screen its own copy.
  test('annual CPA topics read as the generic topic', () {
    for (final topic in [
      CpaTopic.platformNondefinitiveAnnual,
      CpaTopic.umaForMinimumWage,
      CpaTopic.deductionCapResicoIncome,
      CpaTopic.salaryObligationBasis,
      CpaTopic.cuRounding,
      CpaTopic.resicoPmPtuBase,
      CpaTopic.resicoPfReliefRenewal,
      CpaTopic.ptuRate,
    ]) {
      expect(cpaTopicsText(l10n, [topic]), l10n.cpaTopicOther);
    }
  });

  test('each topic is named once', () {
    expect(
      cpaTopicsText(l10n, [
        CpaTopic.withholdingRuleUnverified,
        CpaTopic.withheldIvaCreditTiming,
      ]),
      l10n.cpaTopicOther,
    );
  });
}
