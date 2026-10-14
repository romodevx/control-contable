import 'package:control_contable/features/taxpayers/domain/decimal_rate.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('parseDecimalRate (es-MX input, canonical DecimalRate out)', () {
    for (final (input, expected) in [
      ('0.25', '0.25'),
      ('.25', '0.25'),
      ('0,25', '0.25'),
      (' 0.250 ', '0.25'),
      ('1', '1'),
      ('1.', '1'),
      ('007.5', '7.5'),
      ('0.000001', '0.000001'),
      ('999.999999', '999.999999'),
      ('0', '0'),
      ('0.0', '0'),
    ]) {
      test('"$input" → $expected', () {
        expect(parseDecimalRate(input), expected);
      });
    }

    for (final input in [
      '',
      'abc',
      '-0.1',
      '1.2.3',
      '1,2.3',
      '0.0000001',
      '1000',
      '1e-3',
      '.',
    ]) {
      test('"$input" is not a rate', () {
        expect(parseDecimalRate(input), isNull);
      });
    }
  });

  group('compareDecimalRates (exact, never a float)', () {
    for (final (a, b, sign) in [
      ('0.1', '0.10', 0),
      ('0.1', '0.2', -1),
      ('1', '0.999999', 1),
      ('0.300001', '0.3', 1),
    ]) {
      test('$a vs $b → $sign', () {
        expect(compareDecimalRates(a, b).sign, sign);
      });
    }
  });

  group('isRateWithin', () {
    test('bounds are inclusive and optional', () {
      expect(isRateWithin('0.1', min: '0.1', max: '0.5'), isTrue);
      expect(isRateWithin('0.5', min: '0.1', max: '0.5'), isTrue);
      expect(isRateWithin('0.05', min: '0.1', max: '0.5'), isFalse);
      expect(isRateWithin('0.6', min: '0.1', max: '0.5'), isFalse);
      expect(isRateWithin('42'), isTrue);
    });
  });
}
