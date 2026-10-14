import 'package:control_contable/api/generated/models/iva_components.dart';
import 'package:control_contable/features/tax_engine/domain/iva.dart';
import 'package:flutter_test/flutter_test.dart';

IvaComponents _components({
  int charged = 0,
  int creditable = 0,
  int withheld = 0,
  int excluded = 0,
}) => IvaComponents(
  chargedCollectedCents: charged,
  creditablePaidCents: creditable,
  withheldFromTaxpayerCents: withheld,
  definitiveExcludedCents: excluded,
);

void main() {
  test('a regime with no IVA components does not contribute', () {
    expect(contributesIva(null), isFalse);
    expect(contributesIva(_components()), isFalse);
  });

  test('any non-zero component contributes to the month IVA', () {
    expect(contributesIva(_components(charged: 1)), isTrue);
    expect(contributesIva(_components(creditable: 1)), isTrue);
    expect(contributesIva(_components(withheld: 1)), isTrue);
    expect(contributesIva(_components(excluded: 1)), isTrue);
  });
}
