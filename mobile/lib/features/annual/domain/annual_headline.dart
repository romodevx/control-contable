import 'package:control_contable/api/generated/models/annual_estimate.dart';
import 'package:control_contable/api/generated/models/annual_section.dart';
import 'package:control_contable/api/generated/models/annual_tax.dart';

/// A section whose balance is part of the headline, with its ISR.
typedef HeadlineSection = ({AnnualSection section, AnnualTax tax});

/// The sections the server names in `headline_sections`, in that order,
/// each with its ISR (a PM's two sections each on its own, never netted).
/// A named kind with no section or no ISR is left out: the client never
/// invents a balance.
List<HeadlineSection> headlineSections(AnnualEstimate estimate) => [
  for (final kind in estimate.headlineSections)
    for (final section in estimate.sections)
      if (section.kind.json == kind.json)
        if (section.tax case final tax?) (section: section, tax: tax),
];
