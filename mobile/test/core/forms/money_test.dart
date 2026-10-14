import 'package:control_contable/core/forms/money.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('parseCentavos accepts what people type in es-MX', () {
    const cases = <String, int>{
      '1,234.50': 123450,
      '1234.5': 123450,
      '1234': 123400,
      r'$1,234.50': 123450,
      r'$ 1,234.50': 123450,
      ' 1,234.50 ': 123450,
      '0.05': 5,
      '.5': 50,
      '12,345,678.9': 1234567890,
      '0': 0,
    };
    cases.forEach((input, centavos) {
      test('"$input" -> $centavos', () {
        expect(parseCentavos(input), centavos);
      });
    });
  });

  group('parseCentavos rejects what is not an amount', () {
    const rejected = [
      '',
      ' ',
      r'$',
      '.',
      'abc',
      '-5',
      '1.234,50',
      '1,23',
      '12,34,567',
      '1.234',
      '1.2.3',
      '1,,234',
      ',123',
      '1234.',
      '99999999999999',
    ];
    for (final input in rejected) {
      test('"$input" -> null', () => expect(parseCentavos(input), isNull));
    }
  });

  group('formatCentavos writes the es-MX canonical form', () {
    const cases = <int, String>{
      123450: '1,234.50',
      5: '0.05',
      0: '0.00',
      1234567890: '12,345,678.90',
      999999999999999: '9,999,999,999,999.99',
    };
    cases.forEach((centavos, text) {
      test('$centavos -> "$text"', () {
        expect(formatCentavos(centavos), text);
      });
    });
  });

  group('formatPesos writes an es-MX currency amount with a real minus', () {
    const cases = <int, String>{
      1160000: r'$11,600.00',
      5: r'$0.05',
      0: r'$0.00',
      -206667: r'−$2,066.67',
    };
    cases.forEach((centavos, text) {
      test('$centavos -> "$text"', () {
        expect(formatPesos(centavos), text);
      });
    });
  });
}
