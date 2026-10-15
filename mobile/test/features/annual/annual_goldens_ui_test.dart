import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'annual_fixtures.dart';
import 'pump_annual.dart';

// ANN-23 QA: the figures the annual screen shows for each golden case
// AN-01..AN-08 (MODULE-annual §3), fed with the case's expected section as
// the server sends it. The values are copied from
// backend/internal/taxengine/testdata/annual_golden/{pf,pm}/AN-*.yaml, the
// CI hard gate the engine is held to; this pins that the screen renders
// them to the centavo, with the balance kind in words and the deadline of
// the person type. Before this file only AN-01 (headline), the AN-02
// amount, AN-03 (a) and AN-06 were rendered.

/// One golden case as the screen receives it.
typedef _Golden = ({
  String id,
  String personType,
  String kind,
  List<String> regimes,
  Map<String, Object?> tax,
  String balanceWords,
  String balance,
  String deadline,
  List<(String label, String amount)> lines,
});

const String _pfDeadline = 'Vence el 30 de abril de 2027';
const String _pmDeadline = 'Vence el 31 de marzo de 2027';

final List<_Golden> _goldens = [
  (
    id: 'AN-01',
    personType: 'fisica',
    kind: 'pf_art152',
    regimes: ['612'],
    tax: annualTaxJson(),
    balanceWords: 'Saldo a favor',
    balance: r'$9,408.00',
    deadline: _pfDeadline,
    lines: [
      ('Base gravable', r'$460,000.00'),
      ('ISR del ejercicio', r'$76,365.86'),
    ],
  ),
  (
    id: 'AN-02',
    personType: 'fisica',
    kind: 'pf_art152',
    regimes: ['605', '606', '612'],
    tax: annualTaxJson(
      accumulableCents: 76000000,
      personalDeductionsCents: 16790000,
      baseCents: 59210000,
      isrCents: 10743578,
      paymentsCents: 3800000,
      withholdingsCents: 1500000,
      salaryIsrCents: 5423591,
      balanceKind: 'a_cargo',
      balanceCents: 19987,
    ),
    balanceWords: 'Saldo a cargo',
    balance: r'$199.87',
    deadline: _pfDeadline,
    lines: [
      ('Deducciones personales', r'$167,900.00'),
      ('ISR del ejercicio', r'$107,435.78'),
      ('ISR retenido por sueldos', r'$54,235.91'),
    ],
  ),
  (
    id: 'AN-04',
    personType: 'fisica',
    kind: 'pf_art152',
    regimes: ['625'],
    tax: annualTaxJson(
      accumulableCents: 18000000,
      personalDeductionsCents: 0,
      baseCents: 18000000,
      isrCents: 1683381,
      paymentsCents: 0,
      withholdingsCents: 1200000,
      balanceKind: 'a_cargo',
      balanceCents: 483381,
    ),
    balanceWords: 'Saldo a cargo',
    balance: r'$4,833.81',
    deadline: _pfDeadline,
    lines: [('ISR del ejercicio', r'$16,833.81')],
  ),
  (
    id: 'AN-05',
    personType: 'fisica',
    kind: 'pf_art152',
    regimes: ['605'],
    tax: annualTaxJson(
      accumulableCents: 45000000,
      personalDeductionsCents: 0,
      baseCents: 45000000,
      isrCents: 7401386,
      paymentsCents: 0,
      withholdingsCents: 0,
      salaryIsrCents: 4800000,
      balanceKind: 'a_cargo',
      balanceCents: 2601386,
    ),
    balanceWords: 'Saldo a cargo',
    balance: r'$26,013.86',
    deadline: _pfDeadline,
    lines: [
      ('ISR del ejercicio', r'$74,013.86'),
      ('ISR retenido por sueldos', r'$48,000.00'),
    ],
  ),
  (
    id: 'AN-06',
    personType: 'moral',
    kind: 'pm_601',
    regimes: ['601'],
    tax: annualTaxJson(
      accumulableCents: 1000000000,
      personalDeductionsCents: null,
      baseCents: 118000000,
      isrCents: 35400000,
      paymentsCents: 33000000,
      withholdingsCents: 200000,
      balanceKind: 'a_cargo',
      balanceCents: 2200000,
    ),
    balanceWords: 'Saldo a cargo',
    balance: r'$22,000.00',
    deadline: _pmDeadline,
    lines: [
      ('Resultado fiscal', r'$1,180,000.00'),
      ('ISR del ejercicio', r'$354,000.00'),
    ],
  ),
  (
    id: 'AN-07',
    personType: 'moral',
    kind: 'pm_resico',
    regimes: ['626'],
    tax: annualTaxJson(
      accumulableCents: 600000000,
      personalDeductionsCents: null,
      baseCents: 115000000,
      isrCents: 34500000,
      paymentsCents: 36000000,
      withholdingsCents: 0,
      balanceCents: 1500000,
    ),
    balanceWords: 'Saldo a favor',
    balance: r'$15,000.00',
    deadline: _pmDeadline,
    lines: [('ISR del ejercicio', r'$345,000.00')],
  ),
];

Finder _section(String kind) => find.byKey(ValueKey('annual-section-$kind'));

final Finder _headline = find.byKey(const ValueKey('annual-headline'));

Future<void> _reveal(WidgetTester tester, Finder finder) async {
  await tester.scrollUntilVisible(
    finder,
    200,
    scrollable: find.byType(Scrollable).first,
  );
  await tester.pumpAndSettle();
}

/// [label]'s figure inside [kind]'s section: the row holding both.
void _expectLine(String kind, String label, String amount) {
  final row = find.ancestor(
    of: find.descendant(of: _section(kind), matching: find.text(label)),
    matching: find.byWidgetPredicate((w) => w is Row || w is Wrap || w is Flex),
  );
  expect(row, findsWidgets, reason: label);
  expect(
    find.descendant(
      of: row.first,
      matching: find.text(amount, findRichText: true),
    ),
    findsOneWidget,
    reason: '$label = $amount',
  );
}

void main() {
  late AnnualHarness harness;

  setUp(() => harness = AnnualHarness());

  for (final golden in _goldens) {
    testWidgets('${golden.id}: ${golden.balanceWords} ${golden.balance}, '
        'its deadline and figures', (tester) async {
      if (golden.personType == 'moral') harness.personType = PersonType.moral;
      final isPm = golden.personType == 'moral';
      await harness.pump(
        tester,
        reply: annualEstimateFrom(
          annualEstimateJson(
            personType: golden.personType,
            headline: [golden.kind],
            deadline: isPm ? '2027-03-31' : '2027-04-30',
            sections: [
              sectionJson(
                kind: golden.kind,
                regimes: golden.regimes,
                obligationReasons: const [],
                deadline: isPm ? '2027-03-31' : '2027-04-30',
                tax: golden.tax,
              ),
            ],
          ),
        ),
      );

      expect(
        find.descendant(
          of: _headline,
          matching: find.text(golden.balanceWords),
        ),
        findsOneWidget,
      );
      expect(
        find.descendant(
          of: _headline,
          matching: find.text(golden.balance, findRichText: true),
        ),
        findsOneWidget,
      );
      expect(
        find.descendant(of: _headline, matching: find.text(golden.deadline)),
        findsOneWidget,
      );
      expect(find.text(annualLegendText), findsOneWidget);

      await _reveal(tester, _section(golden.kind));
      for (final (label, amount) in golden.lines) {
        _expectLine(golden.kind, label, amount);
      }
    });
  }

  testWidgets('AN-08: the global cap applies 45,000 of 80,000 and the ISR '
      'before credits is 31,807.91', (tester) async {
    await harness.pump(
      tester,
      reply: annualEstimateFrom(
        annualEstimateJson(
          sections: [
            sectionJson(
              tax: annualTaxJson(
                accumulableCents: 30000000,
                personalDeductionsCents: 4500000,
                baseCents: 25500000,
                isrCents: 3180791,
                paymentsCents: 0,
                withholdingsCents: 0,
                balanceKind: 'a_cargo',
                balanceCents: 3180791,
              ),
            ),
          ],
        ),
      ),
    );

    await _reveal(tester, _section('pf_art152'));
    _expectLine('pf_art152', 'Deducciones personales', r'$45,000.00');
    _expectLine('pf_art152', 'Base gravable', r'$255,000.00');
    _expectLine('pf_art152', 'ISR del ejercicio', r'$31,807.91');
    expect(
      find.text(r'$80,000.00'),
      findsNothing,
      reason: 'never the registered',
    );
  });

  for (final (id, cents, shown) in [
    ('AN-03 (a)', -132000, r'−$1,320.00'),
    ('AN-03 (b)', 27500, r'$275.00'),
  ]) {
    testWidgets('$id: RESICO PF difference $shown is informative, never a '
        'balance', (tester) async {
      final section = resicoSectionJson();
      final check = Map<String, Object?>.of(
        section['resico_check']! as Map<String, Object?>,
      )..['difference_cents'] = cents;
      await harness.pump(
        tester,
        reply: annualEstimateFrom(
          annualEstimateJson(
            headline: const [],
            sections: [
              {...section, 'resico_check': check},
            ],
          ),
        ),
      );

      await _reveal(tester, _section('pf_resico_check'));
      _expectLine('pf_resico_check', 'Diferencia informativa', shown);
      expect(find.textContaining('Saldo a'), findsNothing);
    });
  }
}
