import 'package:control_contable/features/taxpayers/domain/periods.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('the current period is the Mexico City month', () {
    expect(currentPeriod(DateTime.utc(2026, 10, 13, 18)), '2026-10');
    expect(
      currentPeriod(DateTime.utc(2026, 11, 1, 3)),
      '2026-10',
      reason: '03:00 UTC is still Oct 31 in Mexico City',
    );
  });

  test('the months a regime can end in run newest first, from the current '
      'month back to its start', () {
    expect(endPeriodChoices(start: '2026-08', current: '2026-10'), [
      '2026-10',
      '2026-09',
      '2026-08',
    ]);
    expect(endPeriodChoices(start: '2025-12', current: '2026-01'), [
      '2026-01',
      '2025-12',
    ]);
    expect(endPeriodChoices(start: '2026-11', current: '2026-10'), [
      '2026-10',
    ], reason: 'a start in the future still offers the current month');
  });

  test('a regime can start from January 2026, the first fiscal year with '
      'tax parameters, up to the current month, newest first', () {
    expect(firstStartPeriod, '2026-01');
    expect(startPeriodChoices('2026-10'), [
      '2026-10',
      '2026-09',
      '2026-08',
      '2026-07',
      '2026-06',
      '2026-05',
      '2026-04',
      '2026-03',
      '2026-02',
      '2026-01',
    ]);
    expect(startPeriodChoices('2026-01'), ['2026-01']);
    final nextYear = startPeriodChoices('2027-02');
    expect(nextYear.first, '2027-02');
    expect(nextYear.last, '2026-01');
    expect(nextYear, hasLength(14));
  });

  test('a period reads as its first day', () {
    expect(periodMonth('2026-06'), DateTime(2026, 6));
  });
}
