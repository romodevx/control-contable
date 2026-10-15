import 'package:control_contable/api/generated/models/deduction_beneficiary.dart';
import 'package:control_contable/api/generated/models/deduction_category.dart';
import 'package:control_contable/api/generated/models/donee_type.dart';
import 'package:control_contable/api/generated/models/tuition_level.dart';
import 'package:control_contable/l10n/app_localizations.dart';

/// What [category] is called.
String deductionCategoryLabel(
  DeductionCategory category,
  AppLocalizations l10n,
) => switch (category) {
  DeductionCategory.medical => l10n.annualCategoryMedical,
  DeductionCategory.medicalDisability => l10n.annualCategoryMedicalDisability,
  DeductionCategory.funeral => l10n.annualCategoryFuneral,
  DeductionCategory.donation => l10n.annualCategoryDonation,
  DeductionCategory.mortgageRealInterest => l10n.annualCategoryMortgage,
  DeductionCategory.retirementContribution => l10n.annualCategoryRetirement,
  DeductionCategory.healthInsurance => l10n.annualCategoryHealthInsurance,
  DeductionCategory.schoolTransport => l10n.annualCategorySchoolTransport,
  DeductionCategory.localPayrollTax => l10n.annualCategoryLocalPayrollTax,
  DeductionCategory.tuition => l10n.annualCategoryTuition,
  DeductionCategory.$unknown => l10n.annualCategoryOther,
};

/// For whom a deduction was spent.
String beneficiaryLabel(
  DeductionBeneficiary beneficiary,
  AppLocalizations l10n,
) => switch (beneficiary) {
  DeductionBeneficiary.spouse => l10n.annualBeneficiarySpouse,
  DeductionBeneficiary.ascendant => l10n.annualBeneficiaryAscendant,
  DeductionBeneficiary.descendant => l10n.annualBeneficiaryDescendant,
  DeductionBeneficiary.self ||
  DeductionBeneficiary.$unknown => l10n.annualBeneficiarySelf,
};

/// Who received a donation.
String doneeLabel(DoneeType donee, AppLocalizations l10n) => switch (donee) {
  DoneeType.government => l10n.annualDoneeGovernment,
  DoneeType.authorized || DoneeType.$unknown => l10n.annualDoneeAuthorized,
};

/// A school level of the colegiaturas decree.
String tuitionLevelLabel(TuitionLevel level, AppLocalizations l10n) =>
    switch (level) {
      TuitionLevel.preescolar => l10n.annualLevelPreescolar,
      TuitionLevel.primaria => l10n.annualLevelPrimaria,
      TuitionLevel.secundaria => l10n.annualLevelSecundaria,
      TuitionLevel.profesionalTecnico => l10n.annualLevelProfesionalTecnico,
      TuitionLevel.bachillerato ||
      TuitionLevel.$unknown => l10n.annualLevelBachillerato,
    };
