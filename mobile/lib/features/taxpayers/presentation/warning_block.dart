import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:flutter/material.dart';

/// Width of the block's leading rule.
const double _ruleWidth = 3;

/// A non-blocking warning: a heading and a plain explanation on a tinted
/// block with a leading rule; words carry the meaning, not the color.
class WarningBlock extends StatelessWidget {
  /// A warning titled [title] explaining [body].
  const new({required this.title, required this.body, super.key});

  /// The heading.
  final String title;

  /// The plain explanation.
  final String body;

  @override
  Widget build(BuildContext context) {
    final colors = context.tokens.colors;
    final theme = Theme.of(context);
    return Semantics(
      container: true,
      child: DecoratedBox(
        decoration: BoxDecoration(
          color: colors.legendBg,
          border: Border(
            left: BorderSide(color: colors.legendText, width: _ruleWidth),
          ),
        ),
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.md),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Semantics(
                header: true,
                child: Text(
                  title,
                  style: theme.textTheme.titleMedium?.copyWith(
                    color: colors.legendText,
                  ),
                ),
              ),
              const SizedBox(height: AppSpacing.xxs),
              Text(
                body,
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: colors.legendText,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
