import 'package:control_contable/features/taxpayers/domain/regime_in_effect.dart';
import 'package:flutter_test/flutter_test.dart';

import 'taxpayers_fakes.dart';

void main() {
  const october = '2026-10';

  final cases = <({String name, List<String?> endPeriods, bool expected})>[
    (name: 'no regime at all', endPeriods: [], expected: false),
    (name: 'an open regime', endPeriods: [null], expected: true),
    (name: 'ends this month', endPeriods: ['2026-10'], expected: true),
    (name: 'ended last month', endPeriods: ['2026-09'], expected: false),
    (
      name: 'one ended, one open',
      endPeriods: ['2026-06', null],
      expected: true,
    ),
    (
      name: 'every regime ended',
      endPeriods: ['2026-06', '2026-09'],
      expected: false,
    ),
  ];

  for (final c in cases) {
    test('hasRegimeInEffect: ${c.name}', () {
      final regimes = [
        for (final end in c.endPeriods) heldRegime('626', endPeriod: end),
      ];

      expect(hasRegimeInEffect(regimes, period: october), c.expected);
    });
  }
}
