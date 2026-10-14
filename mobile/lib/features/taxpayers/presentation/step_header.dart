import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/taxpayers/application/onboarding_flow.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Height of the progress rule.
const double _progressHeight = 4;

/// The top of an onboarding step: where the user is ("Paso 2 de 4" and a
/// segmented rule), the step's title (a heading for screen readers), and
/// an optional plain explanation.
class StepHeader extends StatelessWidget {
  /// The header of [flow]'s current step, titled [title].
  const new({required this.flow, required this.title, this.intro, super.key});

  /// The onboarding, for the step number.
  final OnboardingFlow flow;

  /// The step's title.
  final String title;

  /// What the step asks for, in plain words.
  final String? intro;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final number = flow.stepNumber;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (number != null) ...[
          Text(
            AppLocalizations.of(context)
                .onboardingProgress(number, flow.totalSteps),
            style: theme.textTheme.bodySmall,
          ),
          const SizedBox(height: AppSpacing.xs),
          ExcludeSemantics(
            child: _SegmentedProgress(done: number, total: flow.totalSteps),
          ),
          const SizedBox(height: AppSpacing.lg),
        ],
        Semantics(
          header: true,
          child: Text(title, style: theme.textTheme.headlineMedium),
        ),
        if (intro case final intro?) ...[
          const SizedBox(height: AppSpacing.xs),
          Text(intro, style: theme.textTheme.bodyMedium),
        ],
      ],
    );
  }
}

/// One segment per step, filled up to the current one: the rule says
/// where the user is without counting.
class _SegmentedProgress extends StatelessWidget {
  const new({required this.done, required this.total});

  final int done;
  final int total;

  @override
  Widget build(BuildContext context) {
    final colors = context.tokens.colors;
    return Row(
      children: [
        for (var step = 1; step <= total; step++) ...[
          if (step > 1) const SizedBox(width: AppSpacing.xxs),
          Expanded(
            child: AnimatedContainer(
              duration: const Duration(milliseconds: 200),
              curve: Curves.easeOutCubic,
              height: _progressHeight,
              decoration: BoxDecoration(
                color: step <= done ? colors.primary : colors.border,
                borderRadius: BorderRadius.circular(_progressHeight / 2),
              ),
            ),
          ),
        ],
      ],
    );
  }
}
