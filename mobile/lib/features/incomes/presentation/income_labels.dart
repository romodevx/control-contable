import 'package:control_contable/api/generated/models/counterparty_type.dart';
import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/api/generated/models/platform_activity.dart';
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

/// An IVA rate, as the user reads it.
String ivaRateLabel(IvaRate rate, AppLocalizations l10n) => switch (rate) {
  IvaRate.rate16 => l10n.ivaRate16,
  IvaRate.rate8 => l10n.ivaRate8,
  IvaRate.rate0 => l10n.ivaRate0,
  IvaRate.exempt => l10n.ivaRateExempt,
  IvaRate.$unknown => l10n.incomeErrorInvalid,
};

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
