import 'package:control_contable/features/incomes/domain/income_basis.dart';
import 'package:control_contable/features/incomes/domain/income_math.dart';
import 'package:flutter_test/flutter_test.dart';

import '../taxpayers/taxpayers_fakes.dart' show heldRegime;

void main() {
  test('the net is the total with IVA minus what was withheld (US-I1)', () {
    const income = IncomeAmounts(
      amountCents: 1000000,
      ivaCents: 160000,
      isrWithheldCents: 100000,
      ivaWithheldCents: 106667,
    );

    expect(income.totalCents, 1160000);
    expect(income.withheldCents, 206667);
    expect(income.netCents, 953333);
  });

  group('incomeBasisOf: the served income_basis of the held regime '
      '(spec §2.2, ADR-0010)', () {
    final held = [
      heldRegime('601', incomeBasis: IncomeBasis.accrual),
      heldRegime('626'),
      heldRegime('612', incomeBasis: IncomeBasis.$unknown),
    ];

    test('accrual when the server says so (the invoice date)', () {
      expect(incomeBasisOf('601', held), IncomeBasis.accrual);
    });

    test('cash when the server says so (the collection date)', () {
      expect(incomeBasisOf('626', held), IncomeBasis.cash);
    });

    test('a value this app does not know reads as cash', () {
      expect(incomeBasisOf('612', held), IncomeBasis.cash);
    });

    test('a regime the taxpayer does not hold reads as cash', () {
      expect(incomeBasisOf('620', held), IncomeBasis.cash);
    });
  });
}
