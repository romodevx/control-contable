import 'package:control_contable/api/generated/models/payment_method.dart';
import 'package:control_contable/core/fiscal/iva_labels.dart';
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
typedef ExpenseSections = ({bool more, bool withholdings});

/// The open sections: as the user left them ([withholdingsOpen] `null`
/// until toggled: opened when the category usually carries one), and
/// always while a field inside shows an error (collapsed fields are not
/// built, so the error summary could not reach them).
ExpenseSections expenseSections(
  ExpenseFormModel model, {
  required bool moreOpen,
  required bool? withholdingsOpen,
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
    withholdings:
        (withholdingsOpen ?? model.suggestsWithholdings) ||
        hasError([model.isrWithheld, model.ivaWithheld]),
  );
}

/// The fields of the expense page (MODULE-expenses §7): the category with
/// its hint and the mandatory legend, the amount with its live total, the
/// IVA rate, the date, how it was paid (with the cash warnings) and the
/// CFDI; the regime only with a choice; then the collapsed sections.
List<Widget> expenseFormFields(
  BuildContext context, {
  required ExpenseFormModel model,
  required String legend,
  required Map<String, int> categoryUsage,
  required ExpenseSections open,
  required VoidCallback onToggleMore,
  required VoidCallback onToggleWithholdings,
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
    if (model.isMoral)
      FormDisclosure(
        title: l10n.expenseSectionWithholdings,
        help: l10n.expenseWithholdingsHelp,
        isOpen: open.withholdings,
        onToggle: onToggleWithholdings,
        children: [
          AppMoneyField(
            field: model.isrWithheld,
            textInputAction: TextInputAction.next,
          ),
          AppMoneyField(
            field: model.ivaWithheld,
            textInputAction: TextInputAction.done,
          ),
        ],
      ),
  ];
}
