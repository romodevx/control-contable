import 'package:control_contable/features/tax_engine/domain/prior_payment_suggestion.dart';
import 'package:flutter_test/flutter_test.dart';

import 'estimate_fixtures.dart';

void main() {
  final cases = <({String name, Map<String, Object?> regime, int? want})>[
    (
      name: 'a complete regime offers its payable ISR',
      regime: regimeJson('612', isrCents: 15096),
      want: 15096,
    ),
    (
      name: 'a partial regime offers its payable ISR',
      regime: regimeJson('612', status: 'partial', isrCents: 15096),
      want: 15096,
    ),
    (
      name: 'an ISR in favor offers zero, never a negative payment',
      regime: regimeJson('612', isrCents: -5000),
      want: 0,
    ),
    (
      name: 'an incomplete regime offers nothing (it would overstate)',
      regime: regimeJson('612', status: 'incomplete', isrCents: 415096),
      want: null,
    ),
    (
      name: 'an informational regime offers nothing',
      regime: regimeJson('612', status: 'informational'),
      want: null,
    ),
    (
      name: 'a regime absent from that month offers nothing',
      regime: regimeJson('606', isrCents: 15096),
      want: null,
    ),
  ];

  for (final c in cases) {
    test(c.name, () {
      final estimate = estimateFrom(estimateJson(regimes: [c.regime]));

      expect(suggestedPriorPayment(estimate, '612'), c.want);
    });
  }
}
