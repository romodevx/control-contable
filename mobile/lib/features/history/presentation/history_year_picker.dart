import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// The fiscal-year picker: only years with published parameters (2026 ..
/// the current one), newest first (MODULE-history §2.1).
class HistoryYearPicker extends StatelessWidget {
  /// The picker showing [selected] among [years].
  const new({
    required this.years,
    required this.selected,
    required this.onSelected,
    super.key,
  });

  /// The years offered, newest first.
  final List<int> years;

  /// The year shown.
  final int selected;

  /// Called with the year the user picks.
  final ValueChanged<int> onSelected;

  @override
  Widget build(BuildContext context) {
    final label = AppLocalizations.of(context).historyYearLabel;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text(label, style: Theme.of(context).textTheme.titleSmall),
        const SizedBox(height: AppSpacing.xxs),
        Semantics(
          label: label,
          child: DropdownButtonFormField<int>(
            key: const ValueKey('history-year'),
            initialValue: selected,
            isExpanded: true,
            items: [
              for (final year in years)
                DropdownMenuItem(value: year, child: Text('$year')),
            ],
            onChanged: (year) => onSelected(year!),
          ),
        ),
      ],
    );
  }
}
