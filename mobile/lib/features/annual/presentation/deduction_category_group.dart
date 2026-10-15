import 'package:control_contable/api/generated/models/deduction_category.dart';
import 'package:control_contable/api/generated/models/deduction_category_total.dart';
import 'package:control_contable/api/generated/models/personal_deduction.dart';
import 'package:control_contable/core/forms/flexible_date.dart';
import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/annual/domain/deduction_rules.dart';
import 'package:control_contable/features/annual/presentation/deduction_labels.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// One category of the year's deductions: its name, what is registered,
/// what cash made non-deductible, and its rows, each with a delete button.
class DeductionCategoryGroup extends StatelessWidget {
  /// The group of [total] listing [items].
  const new({
    required this.total,
    required this.items,
    required this.onDelete,
    super.key,
  });

  /// The category's registered totals.
  final DeductionCategoryTotal total;

  /// Its deductions, newest payment first.
  final List<PersonalDeduction> items;

  /// Deletes one deduction.
  final ValueChanged<PersonalDeduction> onDelete;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    final muted = context.tokens.colors.muted;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Semantics(
          header: true,
          child: Text(
            deductionCategoryLabel(total.category, l10n),
            style: text.titleMedium,
          ),
        ),
        const SizedBox(height: AppSpacing.xs),
        Text(
          l10n.annualDeductionsGroupRegistered(
            formatPesos(total.registeredCents),
          ),
          style: text.bodyMedium,
        ),
        if (total.excludedCents > 0)
          Text(
            l10n.annualDeductionsGroupExcluded(
              formatPesos(total.excludedCents),
            ),
            style: text.bodyMedium?.copyWith(color: muted),
          ),
        for (final deduction in items)
          _DeductionRow(deduction: deduction, onDelete: onDelete),
      ],
    );
  }
}

class _DeductionRow extends StatelessWidget {
  const new({required this.deduction, required this.onDelete});

  final PersonalDeduction deduction;
  final ValueChanged<PersonalDeduction> onDelete;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final amount = formatPesos(deduction.amountCents);
    return ListTile(
      contentPadding: EdgeInsets.zero,
      title: Text(amount),
      subtitle: Text(_details(l10n).join(' · ')),
      trailing: IconButton(
        tooltip: l10n.annualDeductionsDeleteTooltip(amount),
        icon: const Icon(Icons.delete_outline),
        onPressed: () => onDelete(deduction),
      ),
    );
  }

  List<String> _details(AppLocalizations l10n) => [
    formatDate(parseBusinessDate(deduction.paidOn)),
    if (beneficiaryCategories.contains(deduction.category))
      beneficiaryLabel(deduction.beneficiary, l10n),
    if (deduction.doneeType case final donee?) doneeLabel(donee, l10n),
    ?deduction.studentLabel,
    if (deduction.tuitionLevel case final level?)
      tuitionLevelLabel(level, l10n),
    if (deduction.excluded != null) l10n.annualDeductionsCashExcluded,
  ];
}

/// The deductions of [category] among [items].
List<PersonalDeduction> deductionsOf(
  DeductionCategory category,
  List<PersonalDeduction> items,
) => [
  for (final item in items)
    if (item.category == category) item,
];
