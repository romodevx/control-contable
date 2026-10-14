import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/l10n/app_localizations.dart';

/// An IVA rate, as the user reads it.
String ivaRateLabel(IvaRate rate, AppLocalizations l10n) => switch (rate) {
  IvaRate.rate16 => l10n.ivaRate16,
  IvaRate.rate8 => l10n.ivaRate8,
  IvaRate.rate0 => l10n.ivaRate0,
  IvaRate.exempt => l10n.ivaRateExempt,
  IvaRate.$unknown => l10n.recordErrorInvalid,
};
