import 'package:control_contable/api/generated/models/deductibility.dart';
import 'package:control_contable/api/generated/models/expense_category.dart';
import 'package:control_contable/api/generated/models/payment_method.dart';
import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/features/expenses/domain/expense_rules.dart';
import 'package:flutter_test/flutter_test.dart';

import '../taxpayers/taxpayers_fakes.dart';

const int _limit = 200000;

ExpenseCategory _software({bool defaultDeductible = true}) => ExpenseCategory(
  code: 'software',
  name: 'Software',
  deductibility: Deductibility.deductible,
  defaultDeductible: defaultDeductible,
  note: null,
  legalBasis: null,
);

void main() {
  group('cashRiskOf: MODULE-expenses §2.5, on the total with IVA', () {
    const cases = <(PaymentMethod, String?, int, int?), CashRisk?>{
      (PaymentMethod.cash, 'papeleria', 200000, _limit): null,
      (PaymentMethod.cash, 'papeleria', 200001, _limit): CashRisk.overLimit,
      (PaymentMethod.cardDebit, 'papeleria', 900000, _limit): null,
      (PaymentMethod.cash, 'combustible', 30000, _limit): CashRisk.fuel,
      (PaymentMethod.transfer, 'combustible', 30000, _limit): null,
      // Without the catalog's limit there is nothing to compare with.
      (PaymentMethod.cash, 'papeleria', 900000, null): null,
    };
    cases.forEach((input, expected) {
      test('$input → $expected', () {
        expect(
          cashRiskOf(
            method: input.$1,
            categoryCode: input.$2,
            totalCents: input.$3,
            limitCents: input.$4,
          ),
          expected,
        );
      });
    });
  });

  group('suggestedDeductibleOf: §2.4 default = rule AND CFDI AND NOT cash', () {
    test('the rule alone', () {
      expect(
        suggestedDeductibleOf(
          category: _software(),
          hasCfdi: true,
          cashRisk: null,
        ),
        isTrue,
      );
    });
    test('no CFDI', () {
      expect(
        suggestedDeductibleOf(
          category: _software(),
          hasCfdi: false,
          cashRisk: null,
        ),
        isFalse,
      );
    });
    test('cash over the limit', () {
      expect(
        suggestedDeductibleOf(
          category: _software(),
          hasCfdi: true,
          cashRisk: CashRisk.overLimit,
        ),
        isFalse,
      );
    });
    test('a conservative rule, or no category yet', () {
      expect(
        suggestedDeductibleOf(
          category: _software(defaultDeductible: false),
          hasCfdi: true,
          cashRisk: null,
        ),
        isFalse,
      );
      expect(
        suggestedDeductibleOf(category: null, hasCfdi: true, cashRisk: null),
        isFalse,
      );
    });
  });

  group('hintOf: §2.3 plain-language hint', () {
    test('the catalog hint for a business regime', () {
      expect(
        hintOf(
          Deductibility.partiallyDeductible,
          regime: heldRegime('612'),
          personType: PersonType.fisica,
        ),
        DeductibilityHint.partial,
      );
    });
    test('RESICO PF never lowers ISR; RESICO PM follows the catalog', () {
      expect(
        hintOf(
          Deductibility.isrNotApplicable,
          regime: heldRegime('626'),
          personType: PersonType.fisica,
        ),
        DeductibilityHint.resico,
      );
      expect(
        hintOf(
          Deductibility.deductible,
          regime: heldRegime('626'),
          personType: PersonType.moral,
        ),
        DeductibilityHint.deductible,
      );
    });
    test('606 with the blind 35 % deduction', () {
      expect(
        hintOf(
          Deductibility.isrNotApplicable,
          regime: heldRegime('606', options: {'deduction_mode': 'blind_35'}),
          personType: PersonType.fisica,
        ),
        DeductibilityHint.blind35,
      );
    });
    test('any other regime where ISR does not apply', () {
      expect(
        hintOf(
          Deductibility.isrNotApplicable,
          regime: heldRegime('625'),
          personType: PersonType.fisica,
        ),
        DeductibilityHint.notIsr,
      );
    });
  });
}
