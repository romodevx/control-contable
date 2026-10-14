import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:flutter/material.dart';

/// An empty record month: what is missing, a warm line on why it is worth
/// adding, and the action that adds one.
class RecordEmptyState extends StatelessWidget {
  /// The empty state reading [title] and [body], offering [actionLabel].
  const new({
    required this.title,
    required this.body,
    required this.actionLabel,
    required this.onAction,
    super.key,
  });

  /// What is missing ("Sin ingresos en octubre de 2026").
  final String title;

  /// Why adding one helps.
  final String body;

  /// The action's label.
  final String actionLabel;

  /// Adds a record.
  final VoidCallback onAction;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.all(AppSpacing.gutter),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(title, style: theme.textTheme.titleMedium),
          const SizedBox(height: AppSpacing.xs),
          Text(body, style: theme.textTheme.bodyMedium),
          const SizedBox(height: AppSpacing.lg),
          FilledButton(onPressed: onAction, child: Text(actionLabel)),
        ],
      ),
    );
  }
}
