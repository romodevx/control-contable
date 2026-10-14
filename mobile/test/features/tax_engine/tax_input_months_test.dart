import 'package:control_contable/features/tax_engine/domain/tax_input_months.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('paymentMonths', () {
    test('this year: the months before the current one', () {
      expect(paymentMonths(fiscalYear: 2026, current: '2026-03'), [1, 2]);
    });

    test('January of this year has no prior month', () {
      expect(paymentMonths(fiscalYear: 2026, current: '2026-01'), isEmpty);
    });

    test('a past year: all twelve months', () {
      expect(
        paymentMonths(fiscalYear: 2026, current: '2027-02'),
        List.generate(12, (index) => index + 1),
      );
    });

    test('a stored month is never hidden', () {
      expect(paymentMonths(fiscalYear: 2026, current: '2026-03', stored: {5}), [
        1,
        2,
        5,
      ]);
    });
  });

  group('monthGroups', () {
    test('six or fewer months make one group', () {
      expect(monthGroups([1, 2, 3, 4, 5, 6]), [
        [1, 2, 3, 4, 5, 6],
      ]);
    });

    test('a full year splits into halves', () {
      expect(monthGroups(List.generate(12, (index) => index + 1)), [
        [1, 2, 3, 4, 5, 6],
        [7, 8, 9, 10, 11, 12],
      ]);
    });

    test('the second half keeps only its months', () {
      expect(monthGroups([1, 2, 3, 4, 5, 6, 7, 8]), [
        [1, 2, 3, 4, 5, 6],
        [7, 8],
      ]);
    });
  });
}
