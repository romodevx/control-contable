import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/forms/form_disclosure.dart';
import 'package:control_contable/core/forms/typed_field_controllers.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/tax_engine/application/tax_input_entries.dart';
import 'package:control_contable/features/tax_engine/domain/tax_input_months.dart';
import 'package:control_contable/features/taxpayers/presentation/option_fields.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

/// A `money_monthly` input (the prior provisional payments, OQ-T1): one
/// amount per month already due. Never more than six month fields show at
/// once (ui-forms-and-modals.md): a full year opens one half at a time,
/// and a half holding an error stays open.
class MonthlyPaymentsSection extends StatelessWidget {
  /// The section of [entry] for [fiscalYear]; [openGroup] is the open half
  /// (an index of `monthGroups`), toggled through [onToggleGroup].
  const new({
    required this.entry,
    required this.fiscalYear,
    required this.openGroup,
    required this.onToggleGroup,
    super.key,
  });

  /// The input and its month fields.
  final MonthlyEntry entry;

  /// The fiscal year the months belong to.
  final int fiscalYear;

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
        if (groups.length == 1) _MonthFields(fields: _fieldsOf(groups.single)),
        if (groups.length > 1)
          for (final (index, group) in groups.indexed)
            FormDisclosure(
              title: _rangeTitle(context, l10n, group),
              isOpen: index == openGroup || _hasError(group),
              onToggle: () => onToggleGroup(index),
              children: [_MonthFields(fields: _fieldsOf(group))],
            ),
      ],
    );
  }

  List<MoneyFieldController> _fieldsOf(List<int> months) => [
    for (final month in months) entry.months[month]!,
  ];

  bool _hasError(List<int> months) =>
      _fieldsOf(months).any((field) => field.error != null);

  String _rangeTitle(
    BuildContext context,
    AppLocalizations l10n,
    List<int> months,
  ) => l10n.taxInputsMonthsRange(
    toBeginningOfSentenceCase(monthName(context, fiscalYear, months.first)),
    monthName(context, fiscalYear, months.last),
  );
}

class _MonthFields extends StatelessWidget {
  const new({required this.fields});

  final List<MoneyFieldController> fields;

  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.stretch,
    spacing: AppSpacing.md,
    children: [
      for (final field in fields)
        AppTextField(
          field: field,
          keyboardType: decimalKeyboard,
          textInputAction: TextInputAction.next,
          autofillHints: const [],
          autocorrect: false,
          prefixText: r'$ ',
        ),
    ],
  );
}

/// The es-MX name of [month] ("enero"), for a month field's label.
String monthName(BuildContext context, int fiscalYear, int month) =>
    DateFormat.MMMM(Localizations.localeOf(context).toString())
        .format(DateTime(fiscalYear, month));
