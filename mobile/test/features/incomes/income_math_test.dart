import 'package:control_contable/features/incomes/domain/income_basis.dart';
import 'package:control_contable/features/incomes/domain/income_math.dart';
import 'package:flutter_test/flutter_test.dart';

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

  group('incomeBasisOf: which date books the income (spec §2.2)', () {
    for (final code in ['601', '620', '622', '623', '624']) {
      test('$code is accrual (the invoice date)', () {
        expect(incomeBasisOf(code), IncomeBasis.accrual);
      });
    }
    for (final code in ['612', '606', '625', '626', '603', '610']) {
      test('$code is cash (the collection date)', () {
        expect(incomeBasisOf(code), IncomeBasis.cash);
      });
    }
  });
}
