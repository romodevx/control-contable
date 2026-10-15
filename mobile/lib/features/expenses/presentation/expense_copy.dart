import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/api/generated/models/payment_method.dart';
import 'package:control_contable/api/generated/models/supplier_kind.dart';
import 'package:control_contable/api/generated/models/supplier_regime.dart';
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/api/generated/models/warning_code.dart';
import 'package:control_contable/core/fiscal/withholding_labels.dart';
import 'package:control_contable/core/forms/choice_controllers.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/records/record_save.dart';
import 'package:control_contable/features/expenses/application/expense_form_model.dart';
import 'package:control_contable/features/expenses/domain/expense_rules.dart';
import 'package:control_contable/l10n/app_localizations.dart';

/// The `details[].code` a field can show as "required".
const String _required = 'required';

/// The labels of the expense form, in es-MX.
ExpenseFormCopy expenseFormCopy(AppLocalizations l10n) => (
  category: l10n.expenseFieldCategory,
  amount: l10n.expenseFieldAmount,
  ivaRate: l10n.expenseFieldIvaRate,
  paidOn: l10n.expenseFieldPaidOn,
  paymentMethod: l10n.expenseFieldPaymentMethod,
  hasCfdi: l10n.expenseFieldHasCfdi,
  regime: l10n.expenseFieldRegime,
  ivaAmount: l10n.expenseFieldIvaAmount,
  deductible: l10n.expenseFieldDeductible,
  cfdiUuid: l10n.expenseFieldCfdiUuid,
  cfdiUuidInvalid: l10n.incomeCfdiInvalid,
  description: l10n.expenseFieldDescription,
  supplier: (
    supplierKind: l10n.expenseFieldSupplierKind,
    supplierRegime: l10n.expenseFieldSupplierRegime,
    operationKind: l10n.expenseFieldOperationKind,
    residentialRent: l10n.withholdingResidentialRent,
  ),
);

/// E1's options (MODULE-withholdings §2.2).
List<ChoiceOption> supplierKindOptions(AppLocalizations l10n) => [
  (value: SupplierKind.pm.json!, label: l10n.supplierKindPm),
  (value: SupplierKind.pfWithActivity.json!, label: l10n.supplierKindPf),
  (value: SupplierKind.foreign.json!, label: l10n.supplierKindForeign),
];

/// E2's options: the regimes of a persona física supplier, then "No sé".
List<ChoiceOption> supplierRegimeOptions(AppLocalizations l10n) => [
  (value: SupplierRegime.value612.json!, label: l10n.supplierRegime612),
  (value: SupplierRegime.value606.json!, label: l10n.supplierRegime606),
  (value: SupplierRegime.value626.json!, label: l10n.supplierRegime626),
  (value: SupplierRegime.value625.json!, label: l10n.supplierRegime625),
  (value: SupplierRegime.value621.json!, label: l10n.supplierRegime621),
  (value: SupplierRegime.unknown.json!, label: l10n.supplierRegimeUnknown),
];

/// A payment method, as the user reads it.
String paymentMethodLabel(PaymentMethod method, AppLocalizations l10n) =>
    switch (method) {
      PaymentMethod.cash => l10n.paymentCash,
      PaymentMethod.transfer => l10n.paymentTransfer,
      PaymentMethod.cardDebit => l10n.paymentCardDebit,
      PaymentMethod.cardCredit => l10n.paymentCardCredit,
      PaymentMethod.check => l10n.paymentCheck,
      PaymentMethod.eWallet => l10n.paymentEWallet,
      PaymentMethod.other || PaymentMethod.$unknown => l10n.paymentOther,
    };

/// The lines of a category hint (the second one only for RESICO PF).
List<String> hintLines(DeductibilityHint hint, AppLocalizations l10n) =>
    switch (hint) {
      DeductibilityHint.deductible => [l10n.expenseHintDeductible],
      DeductibilityHint.partial => [l10n.expenseHintPartial],
      DeductibilityHint.notDeductible => [l10n.expenseHintNotDeductible],
      DeductibilityHint.notIsr => [l10n.expenseHintNotIsr],
      DeductibilityHint.resico => [
        l10n.expenseHintResico,
        l10n.expenseHintResicoIva,
      ],
      DeductibilityHint.blind35 => [l10n.expenseHintBlind35],
    };

/// What the cash rule [risk] costs, with the catalog's [limitCents].
String cashRiskText(CashRisk risk, int limitCents, AppLocalizations l10n) =>
    switch (risk) {
      CashRisk.overLimit => l10n.expenseCashOverLimit(formatPesos(limitCents)),
      CashRisk.fuel => l10n.expenseCashFuel,
    };

/// The copy of a server verdict on one field (MODULE-expenses §4). The
/// server's `message` is generic English and is never shown.
FieldError describeExpenseDetail(api.FieldError detail, AppLocalizations l10n) {
  if (detail.code == _required) return const MissingValue();
  return CustomFieldError(switch (detail.code) {
    'unknown_category' => l10n.expenseErrorUnknownCategory,
    'has_cfdi_mismatch' => l10n.expenseErrorHasCfdiMismatch,
    'out_of_range' => l10n.incomeErrorOutOfRange,
    'iva_amount_mismatch' => l10n.incomeErrorIvaMismatch,
    'rate_not_valid_on_date' => l10n.expenseErrorRateNotValid,
    'withholding_not_applicable' => l10n.expenseErrorWithholdingNotApplicable,
    'withholding_exceeds_base' => l10n.incomeErrorWithholdingExceedsBase,
    'regime_not_active' => l10n.incomeErrorRegimeNotActive,
    'supplier_not_applicable' => l10n.expenseErrorSupplierNotApplicable,
    'residential_rent_exempt' => l10n.withholdingResidentialRentHelp,
    'taxpayer_archived' => l10n.settingsArchivedWrite,
    _ => l10n.recordErrorInvalid,
  });
}

/// What a non-blocking warning says. Unknown codes read the server's
/// message, which the contract makes safe to show in es-MX.
String describeExpenseWarning(Warning warning, AppLocalizations l10n) =>
    switch (warning.code) {
      WarningCode.cashOverThreshold => warning.message,
      WarningCode.noCfdi => l10n.expenseNoCfdi,
      WarningCode.notIsrDeductibleInRegime => l10n.expenseWarningNotIsrInRegime,
      WarningCode.categoryNotDeductible =>
        l10n.expenseWarningCategoryNotDeductible,
      WarningCode.duplicateCfdiUuid => l10n.expenseWarningDuplicateCfdi,
      WarningCode.ivaNotCreditableCashPayment => l10n.expenseCashIva,
      WarningCode.provisionalParameters => l10n.expenseWarningProvisional,
      _ => withholdingWarningText(warning, l10n),
    };

/// The form-level message of a save that did not go through.
String expenseSaveNoticeText(RecordSaveNotice notice, AppLocalizations l10n) =>
    switch (notice) {
      RecordSaveNotice.unreachable => l10n.expenseSaveUnreachable,
      RecordSaveNotice.idReused => l10n.expenseSaveIdReused,
      RecordSaveNotice.gone => l10n.expenseSaveGone,
      RecordSaveNotice.failed => l10n.expenseSaveFailed,
    };
