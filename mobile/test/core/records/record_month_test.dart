import 'package:control_contable/core/records/record_month.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('RecordMonth', () {
    test('reads and writes the contract period (YYYY-MM)', () {
      final month = RecordMonth.parse('2026-09');

      expect(month.year, 2026);
      expect(month.month, 9);
      expect(month.period, '2026-09');
      expect(month.firstDay, DateTime(2026, 9));
    });

    test('steps across year boundaries', () {
      expect(RecordMonth.parse('2026-01').previous.period, '2025-12');
      expect(RecordMonth.parse('2025-12').next.period, '2026-01');
    });

    test('the current month is Mexico City’s, not UTC’s', () {
      // 03:00 UTC on Nov 1 is still Oct 31 in Mexico City (UTC−6).
      expect(
        RecordMonth.current(DateTime.utc(2026, 11, 1, 3)).period,
        '2026-10',
      );
      expect(
        RecordMonth.current(DateTime.utc(2026, 11, 1, 7)).period,
        '2026-11',
      );
    });

    test('compares by calendar order and by value', () {
      final sep = RecordMonth.parse('2026-09');

      expect(sep.isBefore(RecordMonth.parse('2026-10')), isTrue);
      expect(sep.isBefore(RecordMonth.parse('2026-09')), isFalse);
      expect(sep, RecordMonth.parse('2026-09'));
      expect(sep.hashCode, RecordMonth.parse('2026-09').hashCode);
    });

    test('the month a calendar date falls in', () {
      expect(RecordMonth.ofDate(DateTime(2026, 9, 30)).period, '2026-09');
    });
  });

  group('business dates', () {
    test('today is the Mexico City calendar date', () {
      expect(
        businessToday(DateTime.utc(2026, 11, 1, 3)),
        DateTime(2026, 10, 31),
      );
    });

    test('a date travels as YYYY-MM-DD and reads back', () {
      expect(businessDate(DateTime(2026, 9, 5)), '2026-09-05');
      expect(parseBusinessDate('2026-09-05'), DateTime(2026, 9, 5));
    });
  });
}
