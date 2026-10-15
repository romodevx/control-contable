import 'package:control_contable/api/generated/models/counterparty_kind.dart';
import 'package:control_contable/api/generated/models/counterparty_type.dart';
import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/api/generated/models/platform_activity.dart';
import 'package:control_contable/core/forms/choice_controllers.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/features/incomes/domain/income_basis.dart';
import 'package:control_contable/l10n/app_localizations.dart';

/// Who paid, as the user reads it.
String counterpartyLabel(CounterpartyType type, AppLocalizations l10n) =>
    switch (type) {
      CounterpartyType.personaFisica => l10n.counterpartyPersonaFisica,
      CounterpartyType.personaMoral => l10n.counterpartyPersonaMoral,
      CounterpartyType.publicoGeneral => l10n.counterpartyPublicoGeneral,
      CounterpartyType.plataforma => l10n.counterpartyPlataforma,
      CounterpartyType.$unknown => l10n.incomeErrorInvalid,
    };

/// Who was billed (Q1), as the user reads it.
String counterpartyKindLabel(CounterpartyKind kind, AppLocalizations l10n) =>
    switch (kind) {
      CounterpartyKind.pm => l10n.counterpartyKindPm,
      CounterpartyKind.govFederal => l10n.counterpartyKindGovFederal,
      CounterpartyKind.govStateMunicipal =>
        l10n.counterpartyKindGovStateMunicipal,
      CounterpartyKind.pfWithActivity => l10n.counterpartyKindPfWithActivity,
      CounterpartyKind.pfNoActivityOrPublic => l10n.counterpartyKindPfPublic,
      CounterpartyKind.foreign ||
      CounterpartyKind.$unknown => l10n.counterpartyKindForeign,
    };

/// Q1's options, those who can withhold first (MODULE-withholdings §2.1).
List<ChoiceOption> counterpartyKindOptions(AppLocalizations l10n) => [
  for (final kind in const [
    CounterpartyKind.pm,
    CounterpartyKind.govFederal,
    CounterpartyKind.govStateMunicipal,
    CounterpartyKind.pfWithActivity,
    CounterpartyKind.pfNoActivityOrPublic,
    CounterpartyKind.foreign,
  ])
    (value: kind.json!, label: counterpartyKindLabel(kind, l10n)),
];

/// A 625 platform activity, as the user reads it.
String activityLabel(PlatformActivity activity, AppLocalizations l10n) =>
    switch (activity) {
      PlatformActivity.transportDelivery => l10n.activityTransportDelivery,
      PlatformActivity.lodging => l10n.activityLodging,
      PlatformActivity.goodsServices => l10n.activityGoodsServices,
      PlatformActivity.$unknown => l10n.incomeErrorInvalid,
    };

/// The date that books [income] in its month (spec §2.2): the invoice date
/// for an accrual regime, else the collection date.
DateTime bookingDateOf(Income income) {
  final booked = incomeBasisOf(income.regimeCode) == IncomeBasis.accrual
      ? income.invoicedOn
      : income.collectedOn;
  return parseBusinessDate(booked ?? income.collectedOn ?? income.invoicedOn!);
}
