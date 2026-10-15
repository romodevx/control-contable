import 'package:control_contable/features/annual/domain/annual_years.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('defaultAnnualYear (OQ-A8: previous year until 30 April)', () {
    final cases = <({String name, DateTime now, int expected})>[
      (
        name: 'October: the current year',
        now: DateTime.utc(2026, 10, 14, 18),
        expected: 2026,
      ),
      (
        name: 'January: the previous year',
        now: DateTime.utc(2028, 1, 2, 18),
        expected: 2027,
      ),
      (
        name: '30 April, late evening in Mexico City: still the previous',
        now: DateTime.utc(2028, 5, 1, 5, 59),
        expected: 2027,
      ),
      (
        name: '1 May in Mexico City: the current year',
        now: DateTime.utc(2028, 5, 1, 6),
        expected: 2028,
      ),
      (
        name: 'before the first year with parameters: 2026',
        now: DateTime.utc(2026, 3, 1, 18),
        expected: 2026,
      ),
    ];
    for (final c in cases) {
      test(c.name, () => expect(defaultAnnualYear(c.now), c.expected));
    }
  });

  group('annualYears', () {
    test('the current year back to 2026, newest first', () {
      expect(annualYears(DateTime.utc(2028, 6, 1, 18)), [2028, 2027, 2026]);
    });

    test('always offers 2026', () {
      expect(annualYears(DateTime.utc(2026, 1, 5, 18)), [2026]);
    });
  });
}
