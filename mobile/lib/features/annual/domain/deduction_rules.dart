import 'package:control_contable/api/generated/models/deduction_category.dart';

/// Most personal deductions a year holds (contract `limit_reached`).
const int deductionsPerYearLimit = 500;

/// Paid in cash, these are stored but never deductible (LISR 151 last
/// paragraph, colegiaturas decree): the form asks how they were paid.
const Set<DeductionCategory> cashRestrictedCategories = {
  DeductionCategory.medical,
  DeductionCategory.medicalDisability,
  DeductionCategory.schoolTransport,
  DeductionCategory.tuition,
};

/// Spent on the taxpayer or a dependant: the form asks for whom.
const Set<DeductionCategory> beneficiaryCategories = {
  DeductionCategory.medical,
  DeductionCategory.medicalDisability,
  DeductionCategory.funeral,
  DeductionCategory.healthInsurance,
  DeductionCategory.schoolTransport,
  DeductionCategory.tuition,
};

/// Proven by an annual constancia (mortgage interest, retirement
/// contributions): dated 31 December by default.
const Set<DeductionCategory> yearEndCategories = {
  DeductionCategory.mortgageRealInterest,
  DeductionCategory.retirementContribution,
};

/// The date a new deduction of [category] starts with in [fiscalYear]:
/// 31 December for annual constancias and past years, else [today].
DateTime defaultPaidOn(
  DeductionCategory? category,
  int fiscalYear,
  DateTime today,
) {
  final yearEnd = DateTime(fiscalYear, 12, 31);
  if (category != null && yearEndCategories.contains(category)) return yearEnd;
  return today.year == fiscalYear ? today : yearEnd;
}
