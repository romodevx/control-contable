import 'package:control_contable/api/generated/models/annual_balance_kind.dart';
import 'package:control_contable/api/generated/models/annual_obligation.dart';
import 'package:control_contable/api/generated/models/annual_section.dart';
import 'package:control_contable/api/generated/models/annual_section_kind.dart';
import 'package:control_contable/api/generated/models/annual_section_status.dart';
import 'package:control_contable/features/tax_engine/presentation/status_chip.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/widgets.dart';

/// The balance kind in words ("Saldo a cargo", "Saldo a favor", "Sin
/// saldo"): never by colour alone. An unknown kind reads as no balance.
String balanceKindText(AppLocalizations l10n, AnnualBalanceKind kind) =>
    switch (kind) {
      AnnualBalanceKind.aCargo => l10n.annualBalanceACargo,
      AnnualBalanceKind.aFavor => l10n.annualBalanceAFavor,
      AnnualBalanceKind.cero ||
      AnnualBalanceKind.$unknown => l10n.annualBalanceCero,
    };

/// What [section] is called: its kind, or the regime of an explanation.
String sectionTitle(AppLocalizations l10n, AnnualSection section) =>
    switch (section.kind) {
      AnnualSectionKind.pfArt152 => l10n.annualSectionPfArt152,
      AnnualSectionKind.pfResicoCheck => l10n.annualSectionResico,
      AnnualSectionKind.pm601 => l10n.annualSectionPm601,
      AnnualSectionKind.pmResico => l10n.annualSectionPmResico,
      AnnualSectionKind.explanation || AnnualSectionKind.$unknown =>
        l10n.estimateRegimeCode(section.regimes.join(', ')),
    };

/// Whether a section must, may or need not be filed, in words; `null` when
/// the server says nothing (an explanation section).
String? obligationText(AppLocalizations l10n, AnnualObligation? obligation) =>
    switch (obligation) {
      AnnualObligation.obligated => l10n.annualObligated,
      AnnualObligation.optional => l10n.annualOptional,
      AnnualObligation.notObligated => l10n.annualNotObligated,
      AnnualObligation.relieved => l10n.annualRelieved,
      AnnualObligation.$unknown || null => null,
    };

/// The chip of a section's [status]: the word, never the colour.
StatusChip annualStatusChip(BuildContext context, AnnualSectionStatus status) {
  final l10n = AppLocalizations.of(context);
  return StatusChip(
    label: switch (status) {
      AnnualSectionStatus.complete => l10n.estimateComplete,
      AnnualSectionStatus.incomplete => l10n.estimateIncomplete,
      AnnualSectionStatus.informational ||
      AnnualSectionStatus.$unknown => l10n.estimateInformational,
    },
  );
}
