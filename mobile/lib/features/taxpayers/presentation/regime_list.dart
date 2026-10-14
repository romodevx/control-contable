import 'package:control_contable/api/generated/models/regime.dart';
import 'package:control_contable/api/generated/models/treatment.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/taxpayers/presentation/onboarding_copy.dart';
import 'package:control_contable/features/taxpayers/presentation/regime_selection_field.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// The support levels, in the order they are listed. A level this build
/// does not know is listed as explanation only (the safest promise).
const List<Treatment> _levels = [
  Treatment.full,
  Treatment.partial,
  Treatment.informational,
];

/// The catalog grouped by support level, each group with its plain
/// explanation, every regime a 44 px selectable row. A partial or
/// informational regime shows its own explanation before it is chosen; a
/// regime incompatible with an earlier choice shows the reason at once.
class RegimeList extends StatelessWidget {
  /// The rows of [catalog] over [selection].
  const new({required this.catalog, required this.selection, super.key});

  /// The person type's catalog.
  final List<Regime> catalog;

  /// The chosen regimes.
  final RegimeSelectionField selection;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final conflicts = selection.conflicts;
    final first = catalog.first;
    final groups = Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        for (final level in _levels)
          if (_regimesOf(level) case final regimes when regimes.isNotEmpty)
            _Group(
              title: _title(level, l10n),
              explanation: _explanation(level, l10n),
              rows: [
                for (final regime in regimes)
                  _RegimeRow(
                    regime: regime,
                    selection: selection,
                    focusNode: regime == first ? selection.focusNode : null,
                    conflict: switch (conflicts[regime.code]) {
                      final c? => l10n.onboardingRegimeConflict(
                        c.code,
                        c.reason,
                      ),
                      null => null,
                    },
                  ),
              ],
            ),
        if (selection.error case final error?)
          Semantics(
            liveRegion: true,
            child: Text(
              describeFieldError(error, l10n),
              style: Theme.of(context).textTheme.bodySmall
                  ?.copyWith(color: context.tokens.colors.error),
            ),
          ),
      ],
    );
    if (!selection.single) return groups;
    final chosen = selection.selected.firstOrNull?.code;
    return RadioGroup<String>(
      groupValue: chosen,
      onChanged: (code) =>
          selection.toggle(catalog.firstWhere((regime) => regime.code == code)),
      child: groups,
    );
  }

  List<Regime> _regimesOf(Treatment level) => [
    for (final regime in catalog)
      if (_levelOf(regime) == level) regime,
  ];

  Treatment _levelOf(Regime regime) => _levels.contains(regime.treatment)
      ? regime.treatment
      : Treatment.informational;

  String _title(Treatment level, AppLocalizations l10n) => switch (level) {
    Treatment.full => l10n.onboardingSupportFullTitle,
    Treatment.partial => l10n.onboardingSupportPartialTitle,
    _ => l10n.onboardingSupportInfoTitle,
  };

  String _explanation(Treatment level, AppLocalizations l10n) =>
      switch (level) {
        Treatment.full => l10n.onboardingSupportFullBody,
        Treatment.partial => l10n.onboardingSupportPartialBody,
        _ => l10n.onboardingSupportInfoBody,
      };
}

class _Group extends StatelessWidget {
  const new({
    required this.title,
    required this.explanation,
    required this.rows,
  });

  final String title;
  final String explanation;
  final List<Widget> rows;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.only(top: AppSpacing.lg),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Semantics(
            header: true,
            child: Text(title, style: theme.textTheme.titleMedium),
          ),
          const SizedBox(height: AppSpacing.xxs),
          Text(explanation, style: theme.textTheme.bodySmall),
          const SizedBox(height: AppSpacing.xs),
          ...rows,
        ],
      ),
    );
  }
}

class _RegimeRow extends StatelessWidget {
  const new({
    required this.regime,
    required this.selection,
    required this.focusNode,
    required this.conflict,
  });

  final Regime regime;
  final RegimeSelectionField selection;
  final FocusNode? focusNode;
  final String? conflict;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final title = Text(regimeTitle(regime), style: theme.textTheme.bodyLarge);
    final subtitle = _subtitle(context);
    if (selection.single) {
      return RadioListTile<String>(
        value: regime.code,
        focusNode: focusNode,
        contentPadding: EdgeInsets.zero,
        title: title,
        subtitle: subtitle,
      );
    }
    return CheckboxListTile(
      value: selection.isChosen(regime),
      focusNode: focusNode,
      isError: conflict != null,
      controlAffinity: ListTileControlAffinity.leading,
      contentPadding: EdgeInsets.zero,
      title: title,
      subtitle: subtitle,
      onChanged: (_) => selection.toggle(regime),
    );
  }

  Widget? _subtitle(BuildContext context) {
    final explanation = regime.explanation;
    final conflict = this.conflict;
    if (explanation == null && conflict == null) return null;
    final theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (explanation != null)
          Text(explanation, style: theme.textTheme.bodySmall),
        // Announced by the regimes step when it appears (not a live
        // region: that would speak it twice where live regions work).
        if (conflict != null)
          Text(
            conflict,
            style: theme.textTheme.bodySmall?.copyWith(
              color: context.tokens.colors.error,
            ),
          ),
      ],
    );
  }
}
