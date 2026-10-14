import 'package:control_contable/features/tax_engine/domain/money_format.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('formatMoney', () {
    test('formats centavos in es-MX with the peso sign', () {
      expect(formatMoney(228429), r'$2,284.29');
      expect(formatMoney(16633333), r'$166,333.33');
      expect(formatMoney(0), r'$0.00');
      expect(formatMoney(5), r'$0.05');
    });

    test('uses a real minus sign for negative amounts', () {
      expect(formatMoney(-92400), r'−$924.00');
    });

    test('never loses a centavo on large amounts (integer arithmetic)', () {
      expect(formatMoney(999999999999999), r'$9,999,999,999,999.99');
    });
  });

  group('spokenMoney', () {
    test('reads pesos and centavos as words a screen reader says right', () {
      expect(spokenMoney(228429), '2,284 pesos con 29 centavos');
    });

    test('omits zero centavos and uses the singular', () {
      expect(spokenMoney(100000), '1,000 pesos');
      expect(spokenMoney(101), '1 peso con 1 centavo');
      expect(spokenMoney(0), '0 pesos');
    });

    test('says "menos" for a negative amount', () {
      expect(spokenMoney(-92400), 'menos 924 pesos');
    });
  });
}
