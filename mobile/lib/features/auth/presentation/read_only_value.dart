import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:flutter/material.dart';

/// A value the user can read but not change, laid out like a field (label
/// above) so it is not mistaken for one: plain text, no input, a [note]
/// saying why. Screen readers hear label, value and note as one node.
class ReadOnlyValue extends StatelessWidget {
  /// [value] under [label], with [note] below.
  const new({
    required this.label,
    required this.value,
    required this.note,
    super.key,
  });

  /// What the value is (*Correo electrónico*).
  final String label;

  /// The value itself.
  final String value;

  /// Why it cannot be changed.
  final String note;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return MergeSemantics(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: theme.textTheme.labelLarge),
          const SizedBox(height: AppSpacing.xxs),
          Text(value, style: theme.textTheme.bodyLarge),
          const SizedBox(height: AppSpacing.xxs),
          Text(
            note,
            style: theme.textTheme.bodySmall?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            ),
          ),
        ],
      ),
    );
  }
}
