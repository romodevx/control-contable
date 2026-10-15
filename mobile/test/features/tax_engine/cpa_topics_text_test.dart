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

  // Contract 0.11.0 appended the annual topics (ADR-0009); the annual
  // estimate (ANN-17) names each in plain words.
  test('annual CPA topics have their own words', () {
    final expected = {
      CpaTopic.platformNondefinitiveAnnual:
          'la anual de plataformas sin pago definitivo',
      CpaTopic.umaForMinimumWage:
          'el uso de la UMA en lugar del salario mínimo',
      CpaTopic.deductionCapResicoIncome:
          'el tope de deducciones sin tus ingresos de RESICO',
      CpaTopic.salaryObligationBasis:
          r'la obligación de declarar por sueldos de más de $400,000',
      CpaTopic.cuRounding: 'el redondeo del coeficiente de utilidad',
      CpaTopic.resicoPmPtuBase: 'la base de PTU en RESICO de personas morales',
      CpaTopic.resicoPfReliefRenewal:
          'la renovación de la facilidad de RESICO en la RMF 2027',
      CpaTopic.ptuRate: 'la tasa de PTU',
    };
    for (final MapEntry(key: topic, value: words) in expected.entries) {
      expect(cpaTopicsText(l10n, [topic]), words, reason: topic.json);
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
