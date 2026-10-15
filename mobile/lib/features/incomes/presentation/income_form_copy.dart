import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/api/generated/models/warning_code.dart';
import 'package:control_contable/core/fiscal/withholding_labels.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/server_field_errors.dart';
import 'package:control_contable/features/incomes/application/income_form_model.dart';
import 'package:control_contable/l10n/app_localizations.dart';

/// The `details[].code` a field can show as "required".
const String _required = 'required';

/// The labels of the new-income form, in es-MX.
IncomeFormCopy incomeFormCopy(AppLocalizations l10n) => (
  collectedOn: l10n.incomeFieldCollectedOn,
  invoicedOn: l10n.incomeFieldInvoicedOn,
  description: l10n.incomeFieldDescription,
  amount: l10n.incomeFieldAmount,
  ivaRate: l10n.incomeFieldIvaRate,
  facts: (
    counterpartyKind: l10n.incomeFieldCounterpartyKind,
    operationKind: l10n.incomeFieldOperationKind,
    residentialRent: l10n.withholdingResidentialRent,
    viaPlatform: l10n.incomeFieldViaPlatform,
  ),
  regime: l10n.incomeFieldRegime,
  activity: l10n.incomeFieldActivity,
  ivaAmount: l10n.incomeFieldIvaAmount,
  cfdiUuid: l10n.incomeFieldCfdiUuid,
  cfdiUuidInvalid: l10n.incomeCfdiInvalid,
);

/// The copy of a server verdict on one field (MODULE-incomes §4). The
/// server's `message` is generic English and is never shown.
FieldError describeIncomeDetail(api.FieldError detail, AppLocalizations l10n) {
  if (detail.code == _required) return const MissingValue();
  return CustomFieldError(switch (detail.code) {
    'out_of_range' => l10n.incomeErrorOutOfRange,
    invisibleCharactersCode => l10n.fieldInvisibleCharacters,
    'iva_amount_mismatch' => l10n.incomeErrorIvaMismatch,
    'rate_not_valid_on_date' => l10n.incomeErrorRateNotValid,
    'withholding_not_applicable' => l10n.incomeErrorWithholdingNotApplicable,
    'withholding_exceeds_base' => l10n.incomeErrorWithholdingExceedsBase,
    'residential_rent_exempt' => l10n.withholdingResidentialRentHelp,
    'regime_not_active' => l10n.incomeErrorRegimeNotActive,
    'activity_not_held' => l10n.incomeErrorActivityNotHeld,
    'taxpayer_archived' => l10n.settingsArchivedWrite,
    _ => l10n.incomeErrorInvalid,
  });
}

/// What a non-blocking warning says. Unknown codes read the server's
/// message, which the contract makes safe to show in es-MX.
String describeIncomeWarning(Warning warning, AppLocalizations l10n) =>
    switch (warning.code) {
      WarningCode.duplicateCfdiUuid => l10n.incomeWarningDuplicateCfdi,
      WarningCode.withholdingDiffersFromRate =>
        l10n.incomeWarningWithholdingDiffers,
      _ => withholdingWarningText(warning, l10n),
    };
