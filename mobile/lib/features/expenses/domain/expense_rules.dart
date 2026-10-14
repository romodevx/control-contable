import 'package:control_contable/api/generated/models/deductibility.dart';
import 'package:control_contable/api/generated/models/expense_category.dart';
import 'package:control_contable/api/generated/models/payment_method.dart';
import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/taxpayer_regime.dart';

/// The category paid electronically or not at all (LISR 27-III).
const String fuelCategory = 'combustible';

/// Categories that usually carry a withholding to the supplier: fees, rent
/// and freight (MODULE-expenses §2.6).
const Set<String> withholdingCategories = {
  'servicios_profesionales',
  'renta_local',
  'fletes',
};

/// RESICO (626): for a persona física expenses never lower ISR.
const String _resico = '626';

/// Arrendamiento (606), whose blind deduction leaves expenses out.
const String _rent = '606';

/// The 606 option and value of the blind 35 % deduction.
const String _deductionMode = 'deduction_mode';
const String _blind35 = 'blind_35';

/// Which cash rule a payment breaks (MODULE-expenses §2.5, ⚠️ CPA).
enum CashRisk {
  /// Cash above the limit, on the total with IVA.
  overLimit,

  /// Fuel paid in cash, whatever the amount.
  fuel,
}

/// The cash rule a payment of [totalCents] (with IVA) by [method] for
/// [categoryCode] breaks, if any; `null` without the catalog's
/// [limitCents] (the rule is dated parameter data, never hard-coded).
CashRisk? cashRiskOf({
  required PaymentMethod method,
  required String? categoryCode,
  required int totalCents,
  required int? limitCents,
}) {
  if (method != PaymentMethod.cash || limitCents == null) return null;
  if (categoryCode == fuelCategory) return CashRisk.fuel;
  return totalCents > limitCents ? CashRisk.overLimit : null;
}

/// The deductibility the server gives without an override (§2.4): the
/// category's default, with a CFDI, not paid in cash against the rule.
bool suggestedDeductibleOf({
  required ExpenseCategory? category,
  required bool hasCfdi,
  required CashRisk? cashRisk,
}) => (category?.defaultDeductible ?? false) && hasCfdi && cashRisk == null;

/// What the form tells the user about a category (§2.3, PRD US-E1).
enum DeductibilityHint {
  /// Usually lowers ISR.
  deductible,

  /// Lowers ISR only in part.
  partial,

  /// Usually does not lower ISR.
  notDeductible,

  /// ISR does not apply in this regime.
  notIsr,

  /// RESICO PF: never lowers ISR, the IVA may still be credited.
  resico,

  /// 606 with the blind 35 % deduction.
  blind35,
}

/// The hint of a category resolved as [deductibility] for [regime] held by
/// a [personType] taxpayer.
DeductibilityHint hintOf(
  Deductibility deductibility, {
  required TaxpayerRegime regime,
  required PersonType personType,
}) => switch (deductibility) {
  Deductibility.deductible => DeductibilityHint.deductible,
  Deductibility.partiallyDeductible => DeductibilityHint.partial,
  Deductibility.notDeductible => DeductibilityHint.notDeductible,
  Deductibility.isrNotApplicable ||
  Deductibility.$unknown => _whyNoIsr(regime, personType),
};

DeductibilityHint _whyNoIsr(TaxpayerRegime regime, PersonType personType) {
  if (regime.regimeCode == _resico && personType == PersonType.fisica) {
    return DeductibilityHint.resico;
  }
  if (regime.regimeCode == _rent &&
      regime.options[_deductionMode] == _blind35) {
    return DeductibilityHint.blind35;
  }
  return DeductibilityHint.notIsr;
}
