import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/core/fiscal/iva_math.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('ivaCentsFor prefills the IVA from the rate, in integer centavos', () {
    const cases = <(int, IvaRate), int>{
      (1000000, IvaRate.rate16): 160000,
      (1000000, IvaRate.rate8): 80000,
      (1000000, IvaRate.rate0): 0,
      (1000000, IvaRate.exempt): 0,
      // 0.16 × 3 = 0.48 → 0; 0.16 × 4 = 0.64 → 1 (half up).
      (3, IvaRate.rate16): 0,
      (4, IvaRate.rate16): 1,
      // Far above any personal amount: no float, no overflow.
      (999999999999999, IvaRate.rate16): 160000000000000,
    };
    cases.forEach((input, expected) {
      test('${input.$1} at ${input.$2.json} -> $expected', () {
        expect(ivaCentsFor(input.$1, input.$2), expected);
      });
    });
  });

  group('withheldCents applies a DecimalRate to its base, half up', () {
    test('10 % ISR of 10,000.00 is 1,000.00', () {
      expect(withheldCents(1000000, '0.1'), 100000);
    });

    test('2/3 of an IVA of 1,600.00 is 1,066.67 (US-I1)', () {
      expect(withheldCents(160000, '0.666667'), 106667);
    });

    test('4 % of 15,000.00 is 600.00 (625 lodging)', () {
      expect(withheldCents(1500000, '0.04'), 60000);
    });

    test('a huge base does not overflow', () {
      expect(withheldCents(999999999999999, '1'), 999999999999999);
    });
  });
}
