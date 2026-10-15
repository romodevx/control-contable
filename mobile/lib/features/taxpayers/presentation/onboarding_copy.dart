import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/regime.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/features/taxpayers/application/onboarding_failures.dart';
import 'package:control_contable/features/taxpayers/application/onboarding_flow.dart';
import 'package:control_contable/features/taxpayers/application/regime_wizard.dart';
import 'package:control_contable/features/taxpayers/domain/rfc.dart';
import 'package:control_contable/l10n/app_localizations.dart';

/// `display_name` length limit (`TaxpayerCreate`).
const int displayNameMaxLength = 80;

/// The client check of an RFC typed for [personType] (the server decides).
FieldError? checkTypedRfc(
  String value,
  PersonType personType,
  AppLocalizations l10n,
) {
  final problem = checkRfc(normalizeRfc(value), personType);
  if (problem == null) return null;
  return CustomFieldError(switch (problem) {
    RfcProblem.length => _lengthMessage(personType, l10n),
    RfcProblem.format => l10n.onboardingRfcFormat,
    RfcProblem.date => l10n.onboardingRfcDate,
    RfcProblem.generic => l10n.onboardingRfcGeneric,
  });
}

/// The field error for the server's RFC verdict.
FieldError describeRfcRejection(
  RfcRejection rejection,
  PersonType personType,
  AppLocalizations l10n,
) => CustomFieldError(switch (rejection) {
  RfcRejection.length => _lengthMessage(personType, l10n),
  RfcRejection.format => l10n.onboardingRfcFormat,
  RfcRejection.date => l10n.onboardingRfcDate,
  RfcRejection.generic => l10n.onboardingRfcGeneric,
  RfcRejection.duplicate => l10n.onboardingRfcDuplicate,
});

/// The client check of a display name.
FieldError? checkDisplayName(String value, AppLocalizations l10n) =>
    value.runes.length > displayNameMaxLength
    ? CustomFieldError(l10n.onboardingDisplayNameTooLong)
    : null;

/// The field error for the server's display name verdict.
FieldError describeDisplayNameRejection(
  DisplayNameRejection rejection,
  AppLocalizations l10n,
) => CustomFieldError(switch (rejection) {
  DisplayNameRejection.invisibleCharacters => l10n.fieldInvisibleCharacters,
  DisplayNameRejection.length => l10n.onboardingDisplayNameTooLong,
});

/// The form-level message for [notice].
String describeNotice(OnboardingNotice notice, AppLocalizations l10n) =>
    switch (notice) {
      OnboardingNotice.unreachable => l10n.onboardingErrorUnreachable,
      OnboardingNotice.rateLimited => l10n.onboardingErrorRateLimited,
      OnboardingNotice.limitReached => l10n.onboardingErrorLimit,
      OnboardingNotice.combinationRejected =>
        l10n.onboardingRegimeCombinationRejected,
      OnboardingNotice.regimeRejected => l10n.onboardingRegimeRejected,
      OnboardingNotice.startPeriodRejected =>
        l10n.onboardingStartPeriodRejected,
      OnboardingNotice.unexpected => l10n.onboardingErrorUnexpected,
    };

/// How a regime reads in lists and headings: `606 · Arrendamiento`.
String regimeTitle(Regime regime) => '${regime.code} · ${regime.name}';

/// The label of an option value the catalog sends as a key (`blind_35`);
/// a value this build has no copy for reads as sent.
String optionValueLabel(String key, String value, AppLocalizations l10n) =>
    switch ((key, value)) {
      (_, 'blind_35') => l10n.optionValueBlind35,
      (_, 'actual') => l10n.optionValueActual,
      (_, 'transport_delivery') => l10n.optionValueTransportDelivery,
      (_, 'lodging') => l10n.optionValueLodging,
      (_, 'goods_services') => l10n.optionValueGoodsServices,
      ('rif_year', _) => l10n.optionValueRifYear(value),
      _ => value,
    };

String _lengthMessage(PersonType personType, AppLocalizations l10n) =>
    personType == PersonType.moral
    ? l10n.onboardingRfcLengthMoral
    : l10n.onboardingRfcLengthFisica;

/// The primary of a wizard's last page.
String finishLabel(RegimeWizardPurpose purpose, AppLocalizations l10n) =>
    switch (purpose) {
      RegimeWizardPurpose.createTaxpayer => l10n.onboardingCreate,
      RegimeWizardPurpose.addRegime => l10n.settingsAddRegime,
    };

/// What the regimes step asks for, in plain words.
String regimesIntro(RegimeWizard flow, AppLocalizations l10n) =>
    switch ((flow.purpose, flow.personType)) {
      (RegimeWizardPurpose.addRegime, _) => l10n.addRegimeIntro,
      (_, PersonType.moral) => l10n.onboardingRegimesIntroMoral,
      _ => l10n.onboardingRegimesIntroFisica,
    };

/// [items] joined as a Spanish list reads: `a`, `a y b`, `a, b y c`.
String joinAsList(List<String> items, AppLocalizations l10n) => items.length < 2
    ? items.join()
    : l10n.listAnd(items.take(items.length - 1).join(', '), items.last);

/// A CPA-review warning in plain words: which regimes were registered
/// together, then the server's reason; the reason alone when the warning
/// names no pair (QA M2-12 D3).
String describeCpaReview(CpaReview review, AppLocalizations l10n) {
  if (review.regimes.length < 2) return review.reason;
  final regimes = [
    for (final regime in review.regimes) '«${regimeTitle(regime)}»',
  ];
  return l10n.onboardingCpaWarningBody(
    joinAsList(regimes, l10n),
    review.reason,
  );
}
