import 'package:control_contable/api/generated/models/annual_estimate.dart';
import 'package:control_contable/api/generated/models/annual_section_kind.dart';

/// Salaries (605): the employer withholds; there is no provisional payment.
const String _salaryRegime = '605';

/// The sections whose regimes pay provisional ISR (`getFiledPayments`
/// refuses the others): explanation sections are left out.
const Set<AnnualSectionKind> _paymentSections = {
  AnnualSectionKind.pfArt152,
  AnnualSectionKind.pfResicoCheck,
  AnnualSectionKind.pm601,
  AnnualSectionKind.pmResico,
};

/// The regimes of [estimate] whose provisional payments can be confirmed,
/// in section order, each once.
List<String> provisionalPaymentRegimes(AnnualEstimate estimate) => {
  for (final section in estimate.sections)
    if (_paymentSections.contains(section.kind))
      for (final regime in section.regimes)
        if (regime != _salaryRegime) regime,
}.toList();
