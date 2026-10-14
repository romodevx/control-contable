import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:flutter/material.dart';

/// Progressive disclosure inside a form: a 44 px header button that shows
/// or hides [children] (what the 80 % case does not need). Screen readers
/// hear it as expanded or collapsed. Collapsed children are not built, so
/// a screen keeps it open while one of them shows an error.
class FormDisclosure extends StatelessWidget {
  /// A section titled [title].
  const new({
    required this.title,
    required this.isOpen,
    required this.onToggle,
    required this.children,
    this.help,
    super.key,
  });

  /// The section's name ("Retenciones").
  final String title;

  /// Whether [children] show.
  final bool isOpen;

  /// Opens or closes the section.
  final VoidCallback onToggle;

  /// A short line under the title while open.
  final String? help;

  /// The section's fields.
  final List<Widget> children;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final help = this.help;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        // One node, so the button itself is announced expanded/collapsed.
        MergeSemantics(
          child: Semantics(
            expanded: isOpen,
            child: TextButton.icon(
              style: TextButton.styleFrom(
                alignment: Alignment.centerLeft,
                minimumSize: const Size.fromHeight(AppSizes.minTarget),
              ),
              icon: Icon(isOpen ? Icons.expand_less : Icons.expand_more),
              label: Text(title),
              onPressed: onToggle,
            ),
          ),
        ),
        if (isOpen) ...[
          if (help != null) ...[
            Text(help, style: theme.textTheme.bodySmall),
            const SizedBox(height: AppSpacing.sm),
          ],
          for (final (index, child) in children.indexed) ...[
            if (index > 0) const SizedBox(height: AppSpacing.lg),
            child,
          ],
        ],
      ],
    );
  }
}
