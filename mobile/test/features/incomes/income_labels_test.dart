import 'package:control_contable/api/generated/models/counterparty_kind.dart';
import 'package:control_contable/features/incomes/presentation/income_labels.dart';
import 'package:control_contable/l10n/app_localizations_es.dart';
import 'package:flutter_test/flutter_test.dart';

import 'incomes_fakes.dart';

void main() {
  final l10n = AppLocalizationsEs();

  // A row without a description is titled by who paid, from the facts
  // since 0.12.0 (WH-11): the same words the removed counterparty_type gave.
  group('counterpartyLabel', () {
    final cases = {
      CounterpartyKind.pm: 'Persona moral',
      CounterpartyKind.govFederal: 'Persona moral',
      CounterpartyKind.govStateMunicipal: 'Persona moral',
      CounterpartyKind.pfWithActivity: 'Persona física',
      CounterpartyKind.pfNoActivityOrPublic: 'Público en general',
      CounterpartyKind.foreign: 'Público en general',
    };
    for (final MapEntry(key: kind, value: label) in cases.entries) {
      test('${kind.json} reads "$label"', () {
        expect(counterpartyLabel(income(counterpartyKind: kind), l10n), label);
      });
    }

    test('an income collected through a platform reads as the platform', () {
      expect(
        counterpartyLabel(income(viaPlatform: true), l10n),
        'Plataforma digital',
      );
    });
  });
}
