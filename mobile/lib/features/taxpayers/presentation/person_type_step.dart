import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/taxpayers/application/onboarding_flow.dart';
import 'package:control_contable/features/taxpayers/presentation/step_header.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Size of the card's leading icon.
const double _iconSize = 28;

/// Border widths of a card: resting and chosen.
const double _restingBorder = 1;
const double _chosenBorder = 2;

/// Step 1 (OD-3): persona física or moral, two large choices explained in
/// plain words. Choosing one moves on: one decision, one tap.
class PersonTypeStep extends StatelessWidget {
  /// The step of [flow].
  const new({required this.flow, super.key});

  /// The onboarding.
  final OnboardingFlow flow;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return SingleChildScrollView(
      padding: const EdgeInsets.all(AppSpacing.gutter),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          StepHeader(
            flow: flow,
            title: l10n.onboardingPersonTypeTitle,
            intro: l10n.onboardingPersonTypeIntro,
          ),
          const SizedBox(height: AppSpacing.xl),
          _ChoiceCard(
            icon: Icons.person_outline,
            title: l10n.onboardingFisicaTitle,
            body: l10n.onboardingFisicaBody,
            isChosen: flow.personType == PersonType.fisica,
            onTap: () => flow.choosePersonType(PersonType.fisica),
          ),
          const SizedBox(height: AppSpacing.md),
          _ChoiceCard(
            icon: Icons.business_outlined,
            title: l10n.onboardingMoralTitle,
            body: l10n.onboardingMoralBody,
            isChosen: flow.personType == PersonType.moral,
            onTap: () => flow.choosePersonType(PersonType.moral),
          ),
        ],
      ),
    );
  }
}

/// A large tappable choice: icon, title and explanation. The chosen one
/// (when coming back) has a heavier primary border and a check, and reads
/// as selected.
class _ChoiceCard extends StatelessWidget {
  const new({
    required this.icon,
    required this.title,
    required this.body,
    required this.isChosen,
    required this.onTap,
  });

  final IconData icon;
  final String title;
  final String body;
  final bool isChosen;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final tokens = context.tokens;
    final colors = tokens.colors;
    final theme = Theme.of(context);
    final shape = RoundedRectangleBorder(
      borderRadius: BorderRadius.circular(tokens.radii.card),
      side: BorderSide(
        color: isChosen ? colors.primary : colors.outline,
        width: isChosen ? _chosenBorder : _restingBorder,
      ),
    );
    return Semantics(
      button: true,
      selected: isChosen,
      container: true,
      child: Material(
        color: colors.surface,
        shape: shape,
        child: InkWell(
          customBorder: shape,
          onTap: onTap,
          child: Padding(
            padding: const EdgeInsets.all(AppSpacing.lg),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Icon(icon, size: _iconSize, color: colors.primary),
                const SizedBox(width: AppSpacing.md),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(title, style: theme.textTheme.titleMedium),
                      const SizedBox(height: AppSpacing.xxs),
                      Text(body, style: theme.textTheme.bodyMedium),
                    ],
                  ),
                ),
                if (isChosen) ...[
                  const SizedBox(width: AppSpacing.xs),
                  Icon(Icons.check_circle, color: colors.primary),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}
