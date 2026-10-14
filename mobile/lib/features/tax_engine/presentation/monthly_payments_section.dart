import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/forms/form_disclosure.dart';
import 'package:control_contable/core/forms/typed_field_controllers.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/tax_engine/application/tax_input_entries.dart';
import 'package:control_contable/features/tax_engine/application/tax_inputs.dart';
import 'package:control_contable/features/tax_engine/domain/tax_input_months.dart';
import 'package:control_contable/features/tax_engine/presentation/prior_payment_suggestion_line.dart';
import 'package:control_contable/features/taxpayers/domain/periods.dart';
import 'package:control_contable/features/taxpayers/presentation/option_fields.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

/// A `money_monthly` input (the prior provisional payments, OQ-T1): one
/// amount per month already due. Never more than six month fields show at
/// once (ui-forms-and-modals.md): a full year opens one half at a time,
/// and a half holding an error stays open. An empty month offers the app's
/// own estimate of it (MODULE-tax-engine §8), saved only when used.
class MonthlyPaymentsSection extends StatelessWidget {
  /// The section of [entry] for [query]; [openGroup] is the open half
  /// (an index of `monthGroups`), toggled through [onToggleGroup].
  const new({
    required this.entry,
    required this.query,
    required this.openGroup,
    required this.onToggleGroup,
    super.key,
  });

  /// The input and its month fields.
  final MonthlyEntry entry;

  /// Whose inputs: taxpayer, regime and the fiscal year of the months.
  final TaxInputsQuery query;

  /// The half the user opened, if any.
  final int? openGroup;

  /// Opens the half at an index, or closes it when it is the open one.
  final ValueChanged<int> onToggleGroup;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    final groups = monthGroups(entry.months.keys.toList());
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      spacing: AppSpacing.xs,
      children: [
        Semantics(
          header: true,
          child: Text(entry.def.label, style: text.titleSmall),
        ),
        Text(l10n.taxInputsMonthlyHelp, style: text.bodyMedium),
        if (groups.isEmpty) Text(l10n.taxInputsNoMonthsYet),
        if (groups.length == 1)
          _MonthFields(query: query, fields: _fieldsOf(groups.single)),
        if (groups.length > 1)
          for (final (index, group) in groups.indexed)
            FormDisclosure(
              title: _rangeTitle(context, l10n, group),
              isOpen: index == openGroup || _hasError(group),
              onToggle: () => onToggleGroup(index),
              children: [_MonthFields(query: query, fields: _fieldsOf(group))],
            ),
      ],
    );
  }

  Map<int, MoneyFieldController> _fieldsOf(List<int> months) => {
    for (final month in months) month: entry.months[month]!,
  };

  bool _hasError(List<int> months) =>
      _fieldsOf(months).values.any((field) => field.error != null);

  String _rangeTitle(
    BuildContext context,
    AppLocalizations l10n,
    List<int> months,
  ) => l10n.taxInputsMonthsRange(
    toBeginningOfSentenceCase(
      monthName(context, query.fiscalYear, months.first),
    ),
    monthName(context, query.fiscalYear, months.last),
  );
}

class _MonthFields extends StatelessWidget {
  const new({required this.query, required this.fields});

  final TaxInputsQuery query;

  /// Month (1-12) → its amount.
  final Map<int, MoneyFieldController> fields;

  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.stretch,
    spacing: AppSpacing.md,
    children: [
      for (final MapEntry(key: month, value: field) in fields.entries)
        Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            AppTextField(
              field: field,
              keyboardType: decimalKeyboard,
              textInputAction: TextInputAction.next,
              autofillHints: const [],
              autocorrect: false,
              prefixText: r'$ ',
            ),
            PriorPaymentSuggestionLine(
              field: field,
              query: (
                taxpayerId: query.taxpayerId,
                regime: query.regime,
                period: periodOf(query.fiscalYear, month),
              ),
            ),
          ],
        ),
    ],
  );
}

/// The es-MX name of [month] ("enero"), for a month field's label.
String monthName(BuildContext context, int fiscalYear, int month) =>
    DateFormat.MMMM(Localizations.localeOf(context).toString())
        .format(DateTime(fiscalYear, month));
