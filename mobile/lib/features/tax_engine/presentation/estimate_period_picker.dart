import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/tax_engine/domain/estimate_periods.dart';
import 'package:control_contable/features/taxpayers/domain/periods.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// The month picker: the current month (default) back to January 2026,
/// newest first. Future and pre-2026 months are not offered, and an
/// always-visible line says why (MODULE-tax-engine §2.10).
class EstimatePeriodPicker extends StatelessWidget {
  /// The picker showing [selected] among the months up to [current].
  const new({
    required this.current,
    required this.selected,
    required this.onSelected,
    super.key,
  });

  /// The still-open month (America/Mexico_City).
  final String current;

  /// The month shown.
  final String selected;

  /// Called with the month the user picks.
  final ValueChanged<String> onSelected;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final months = MaterialLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    final muted = text.bodyMedium?.copyWith(color: context.tokens.colors.muted);
    final periods = estimablePeriods(current: current);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (periods.contains(selected)) ...[
          Text(l10n.estimatePeriodLabel, style: text.titleSmall),
          const SizedBox(height: AppSpacing.xxs),
          Semantics(
            label: l10n.estimatePeriodLabel,
            child: DropdownButtonFormField<String>(
              key: const ValueKey('estimate-period'),
              initialValue: selected,
              isExpanded: true,
              items: [
                for (final period in periods)
                  DropdownMenuItem(
                    value: period,
                    child: Text(months.formatMonthYear(periodMonth(period))),
                  ),
              ],
              onChanged: (period) => onSelected(period!),
            ),
          ),
          const SizedBox(height: AppSpacing.xxs),
        ],
        Text(l10n.estimatePeriodHelp, style: muted),
        if (selected == current) ...[
          const SizedBox(height: AppSpacing.xxs),
          Text(l10n.estimateCurrentMonth, style: text.bodyMedium),
        ],
      ],
    );
  }
}
