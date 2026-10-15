import 'package:control_contable/api/generated/models/payment_method.dart';
import 'package:control_contable/core/fiscal/iva_labels.dart';
import 'package:control_contable/core/fiscal/withholding_labels.dart';
import 'package:control_contable/core/forms/app_choice_fields.dart';
import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/forms/form_disclosure.dart';
import 'package:control_contable/core/forms/form_field_model.dart';
import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/forms/typed_fields.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/estimate_legend.dart';
import 'package:control_contable/features/expenses/application/expense_form_model.dart';
import 'package:control_contable/features/expenses/presentation/expense_category_field.dart';
import 'package:control_contable/features/expenses/presentation/expense_copy.dart';
import 'package:control_contable/features/expenses/presentation/expense_form_widgets.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

/// Longest description the contract accepts (`RecordDescription`).
const int _descriptionMaxLength = 200;

/// Which collapsed sections are open.
typedef ExpenseSections = ({bool more, bool supplier});

/// The open sections: as the user left them ([supplierOpen] `null` until
/// toggled: opened when the supplier must be chosen or is not a company),
/// and always while a field inside shows an error (collapsed fields are
/// not built, so the error summary could not reach them).
ExpenseSections expenseSections(
  ExpenseFormModel model, {
  required bool moreOpen,
  required bool? supplierOpen,
}) {
  bool hasError(List<AppFormFieldModel> fields) =>
      fields.any((field) => field.error != null);
  return (
    more:
        moreOpen ||
        hasError([
          model.ivaAmount,
          model.deductible,
          model.cfdiUuid,
          model.description,
        ]),
    supplier:
        (supplierOpen ?? model.supplier.opensItself) ||
        hasError(model.supplier.fieldsByPath.values.toList()),
  );
}

/// The fields of the expense page (MODULE-expenses §7, WH-12): the
/// category with its hint and the mandatory legend, "Tu proveedor" and
/// the derived [card] of a persona moral, the amount with its live total,
/// the IVA rate, the date, how it was paid (with the cash warnings) and
/// the CFDI; the regime only with a choice; then "Más detalles".
List<Widget> expenseFormFields(
  BuildContext context, {
  required ExpenseFormModel model,
  required String legend,
  required Map<String, int> categoryUsage,
  required ExpenseSections open,
  required VoidCallback onToggleMore,
  required VoidCallback onToggleSupplier,
  Widget? card,
}) {
  final l10n = AppLocalizations.of(context);
  final chosen = model.chosenCategory;
  final hint = model.hint;
  final risk = model.cashRisk;
  final limit = model.catalog?.cashPaymentLimitCents;
  return [
    ExpenseCategoryField(
      field: model.category,
      categories: model.categoriesByUse(categoryUsage),
    ),
    if (chosen != null && hint != null)
      CategoryHint(
        key: const ValueKey('category-hint'),
        lines: hintLines(hint, l10n),
        note: chosen.note,
        basis: chosen.legalBasis,
      ),
    EstimateLegend(text: legend),
    if (model.isMoral) ...[
      _supplierSection(
        model,
        l10n,
        isOpen: open.supplier,
        onToggle: onToggleSupplier,
      ),
      ?card,
    ],
    Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        AppMoneyField(
          field: model.amount,
          textInputAction: TextInputAction.done,
        ),
        const SizedBox(height: AppSpacing.xs),
        Text(
          l10n.expenseSummaryTotal(formatPesos(model.totalCents)),
          style: Theme.of(context).textTheme.bodyMedium,
        ),
      ],
    ),
    if (model.asksIvaRate)
      AppChoiceField(
        field: model.ivaRate,
        options: [
          for (final rate in model.ivaRatesOffered)
            (value: rate.json!, label: ivaRateLabel(rate, l10n)),
        ],
      ),
    AppDateField(field: model.paidOn, textInputAction: TextInputAction.next),
    InlineWarnings(
      key: const ValueKey('payment-warnings'),
      field: AppChoiceField(
        field: model.paymentMethod,
        options: [
          for (final method in PaymentMethod.$valuesDefined)
            (value: method.json!, label: paymentMethodLabel(method, l10n)),
        ],
      ),
      warnings: [
        if (risk != null && limit != null) ...[
          cashRiskText(risk, limit, l10n),
          l10n.expenseCashIva,
        ],
      ],
    ),
    InlineWarnings(
      key: const ValueKey('cfdi-warnings'),
      field: AppChoiceField(
        field: model.hasCfdi,
        options: [
          (value: yes, label: l10n.expenseHasCfdiYes),
          (value: no, label: l10n.expenseHasCfdiNo),
        ],
      ),
      warnings: [if (!model.withCfdi) l10n.expenseNoCfdi],
    ),
    if (model.regimes.length > 1)
      AppChoiceField(
        field: model.regime,
        options: [
          for (final regime in model.regimes)
            (value: regime.regimeCode, label: regime.name),
        ],
      ),
    FormDisclosure(
      title: l10n.expenseSectionMore,
      isOpen: open.more,
      onToggle: onToggleMore,
      children: [
        AppMoneyField(
          field: model.ivaAmount,
          textInputAction: TextInputAction.next,
        ),
        DeductibleSwitch(model: model),
        if (model.withCfdi)
          AppTextField(
            field: model.cfdiUuid,
            keyboardType: TextInputType.visiblePassword,
            textInputAction: TextInputAction.next,
            autofillHints: const [],
            autocorrect: false,
            textCapitalization: TextCapitalization.characters,
          ),
        AppTextField(
          field: model.description,
          keyboardType: TextInputType.text,
          textInputAction: TextInputAction.done,
          autofillHints: const [],
          textCapitalization: TextCapitalization.sentences,
          inputFormatters: [
            LengthLimitingTextInputFormatter(_descriptionMaxLength),
          ],
        ),
      ],
    ),
  ];
}

/// "Tu proveedor" (E1–E4): each question only when its answer can change
/// what is withheld.
Widget _supplierSection(
  ExpenseFormModel model,
  AppLocalizations l10n, {
  required bool isOpen,
  required VoidCallback onToggle,
}) {
  final supplier = model.supplier;
  return FormDisclosure(
    title: l10n.expenseSectionSupplier,
    isOpen: isOpen,
    onToggle: onToggle,
    children: [
      AppChoiceField(
        field: supplier.supplierKind,
        options: supplierKindOptions(l10n),
      ),
      if (supplier.asksRegime)
        AppChoiceField(
          field: supplier.supplierRegime,
          helperText: l10n.expenseSupplierRegimeHelp,
          options: supplierRegimeOptions(l10n),
        ),
      if (supplier.asksOperation)
        AppChoiceField(
          field: supplier.operationKind,
          options: operationKindOptions(l10n),
        ),
      if (supplier.asksResidentialRent)
        AppChoiceField(
          field: supplier.residentialRent,
          helperText: l10n.withholdingResidentialRentHelp,
          options: yesNoOptions(l10n),
        ),
    ],
  );
}
