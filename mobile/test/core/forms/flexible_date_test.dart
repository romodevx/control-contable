import 'package:control_contable/core/forms/flexible_date.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('parseFlexibleDate reads day-month-year (es-MX) and ISO', () {
    final cases = <String, DateTime>{
      '31/12/2026': DateTime(2026, 12, 31),
      '31-12-2026': DateTime(2026, 12, 31),
      '31.12.2026': DateTime(2026, 12, 31),
      '31 12 2026': DateTime(2026, 12, 31),
      '1/2/2026': DateTime(2026, 2),
      '01/02/26': DateTime(2026, 2),
      '31122026': DateTime(2026, 12, 31),
      '311226': DateTime(2026, 12, 31),
      '2026-12-31': DateTime(2026, 12, 31),
      ' 29/02/2028 ': DateTime(2028, 2, 29),
    };
    for (final MapEntry(key: input, value: date) in cases.entries) {
      test('"$input" -> $date', () {
        expect(parseFlexibleDate(input), date);
      });
    }
  });

  group('parseFlexibleDate rejects impossible or ambiguous input', () {
    const rejected = [
      '',
      'mañana',
      '31/02/2026',
      '29/02/2027',
      '00/01/2026',
      '1/13/2026',
      '12/31/2026',
      '1/2',
      '1234567',
      '2026-13-01',
      '31/12/02026',
    ];
    for (final input in rejected) {
      test('"$input" -> null', () => expect(parseFlexibleDate(input), isNull));
    }
  });

  test('formatDate writes dd/MM/yyyy', () {
    expect(formatDate(DateTime(2026, 2)), '01/02/2026');
    expect(formatDate(DateTime(2026, 12, 31)), '31/12/2026');
  });
}
