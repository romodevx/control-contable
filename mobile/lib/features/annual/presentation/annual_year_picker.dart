import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/core/widgets/captioned_picker.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// The fiscal-year picker: only years with transcribed parameters (2026 ..
/// the current one), newest first, and an always-visible line saying why
/// and which year opens by default (MODULE-annual OQ-A8).
class AnnualYearPicker extends StatelessWidget {
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
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    final label = l10n.annualYearLabel;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        CaptionedPicker(
          caption: label,
          style: text.titleSmall,
          picker: DropdownButtonFormField<int>(
            key: const ValueKey('annual-year'),
            initialValue: selected,
            isExpanded: true,
            items: [
              for (final year in years)
                DropdownMenuItem(value: year, child: Text('$year')),
            ],
            onChanged: (year) => onSelected(year!),
          ),
        ),
        const SizedBox(height: AppSpacing.xxs),
        Text(
          l10n.annualYearHelp,
          style: text.bodyMedium?.copyWith(color: context.tokens.colors.muted),
        ),
      ],
    );
  }
}
