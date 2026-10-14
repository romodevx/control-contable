import 'package:control_contable/core/forms/choice_controllers.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// One-of filter chips over a record list: *Todos* plus one chip per
/// option. Renders nothing with fewer than two options (nothing to narrow).
class RecordFilterChips extends StatelessWidget {
  /// Chips labeled [label] over [options]; [selected] `null` is *Todos*.
  const new({
    required this.label,
    required this.options,
    required this.selected,
    required this.onSelected,
    super.key,
  });

  /// What the chips filter by (read by screen readers).
  final String label;

  /// The values a list can be narrowed to.
  final List<ChoiceOption> options;

  /// The chosen value, or `null` for all.
  final String? selected;

  /// Called with the chosen value (`null` for all).
  final ValueChanged<String?> onSelected;

  @override
  Widget build(BuildContext context) {
    if (options.length < 2) return const SizedBox.shrink();
    final all = AppLocalizations.of(context).recordsFilterAll;
    return Semantics(
      container: true,
      label: label,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: AppSpacing.gutter),
        child: Wrap(
          spacing: AppSpacing.xs,
          runSpacing: AppSpacing.xs,
          children: [
            _chip(all, isSelected: selected == null),
            for (final option in options)
              _chip(
                option.label,
                isSelected: selected == option.value,
                value: option.value,
              ),
          ],
        ),
      ),
    );
  }

  Widget _chip(String text, {required bool isSelected, String? value}) =>
      ChoiceChip(
        label: Text(text),
        selected: isSelected,
        materialTapTargetSize: MaterialTapTargetSize.padded,
        onSelected: (_) => onSelected(value),
      );
}
