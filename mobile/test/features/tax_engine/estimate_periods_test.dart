import 'package:control_contable/features/tax_engine/domain/estimate_periods.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('estimablePeriods', () {
    test('runs from the current month back to January 2026, newest first', () {
      expect(estimablePeriods(current: '2026-03'), [
        '2026-03',
        '2026-02',
        '2026-01',
      ]);
    });

    test('crosses into a new fiscal year without skipping months', () {
      final periods = estimablePeriods(current: '2027-02');

      expect(periods.first, '2027-02');
      expect(periods[1], '2027-01');
      expect(periods[2], '2026-12');
      expect(periods.last, firstEstimablePeriod);
      expect(periods, hasLength(14));
    });

    test('is empty before the first fiscal year with parameters', () {
      expect(estimablePeriods(current: '2025-12'), isEmpty);
    });
  });

  test('fiscalYearOf reads the year of a period', () {
    expect(fiscalYearOf('2026-11'), 2026);
  });
}
