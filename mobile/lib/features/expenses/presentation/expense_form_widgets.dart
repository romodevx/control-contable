import 'dart:async';

import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/expenses/application/expense_form_model.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';

/// Politely announces each of [messages] not in [before], after the frame
/// that shows them: a live region alone is silent on VoiceOver, and these
/// texts appear without moving focus.
void _announceNew(
  State state,
  List<String> messages, {
  List<String> before = const [],
}) {
  final fresh = messages.where((m) => !before.contains(m)).toList();
  if (fresh.isEmpty) return;
  WidgetsBinding.instance.addPostFrameCallback((_) {
    if (!state.mounted) return;
    final context = state.context;
    for (final message in fresh) {
      unawaited(
        SemanticsService.sendAnnouncement(
          View.of(context),
          message,
          Directionality.of(context),
        ),
      );
    }
  });
}

/// The plain-language hint of the chosen category: [lines], then its
/// [note] and legal [basis] when the catalog has them. The [lines] are
/// announced whenever they change.
class CategoryHint extends StatefulWidget {
  /// The hint reading [lines].
  const new({required this.lines, this.note, this.basis, super.key});

  /// What the category usually means for ISR in this regime.
  final List<String> lines;

  /// The catalog's note ("Sólo una parte es deducible").
  final String? note;

  /// The article-level basis (`LISR 103`).
  final String? basis;

  @override
  State<CategoryHint> createState() => _CategoryHintState();
}

class _CategoryHintState extends State<CategoryHint> {
  @override
  void initState() {
    super.initState();
    _announceNew(this, widget.lines);
  }

  @override
  void didUpdateWidget(CategoryHint oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (!listEquals(widget.lines, oldWidget.lines)) {
      _announceNew(this, widget.lines);
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final l10n = AppLocalizations.of(context);
    final CategoryHint(:lines, :note, :basis) = widget;
    return Semantics(
      container: true,
      liveRegion: true,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          for (final line in lines)
            Text(line, style: theme.textTheme.bodyLarge),
          if (note case final note?) ...[
            const SizedBox(height: AppSpacing.xxs),
            Text(note, style: theme.textTheme.bodyMedium),
          ],
          if (basis != null) ...[
            const SizedBox(height: AppSpacing.xxs),
            Text(
              l10n.expenseHintBasis(basis),
              style: theme.textTheme.bodySmall?.copyWith(
                color: context.tokens.colors.muted,
              ),
            ),
          ],
        ],
      ),
    );
  }
}

/// A field followed by its non-blocking [warnings] (never errors: the save
/// goes through). Each is announced as it appears, with an icon and words,
/// never color alone.
class InlineWarnings extends StatefulWidget {
  /// [field] with [warnings] under it.
  const new({required this.field, required this.warnings, super.key});

  /// The field the warnings speak of.
  final Widget field;

  /// What the current values imply.
  final List<String> warnings;

  @override
  State<InlineWarnings> createState() => _InlineWarningsState();
}

class _InlineWarningsState extends State<InlineWarnings> {
  @override
  void initState() {
    super.initState();
    _announceNew(this, widget.warnings);
  }

  @override
  void didUpdateWidget(InlineWarnings oldWidget) {
    super.didUpdateWidget(oldWidget);
    _announceNew(this, widget.warnings, before: oldWidget.warnings);
  }

  @override
  Widget build(BuildContext context) {
    final style = Theme.of(context).textTheme.bodyMedium;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        widget.field,
        for (final warning in widget.warnings)
          Semantics(
            liveRegion: true,
            container: true,
            child: Padding(
              padding: const EdgeInsets.only(top: AppSpacing.xs),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const ExcludeSemantics(
                    child: Icon(Icons.info_outline, size: 20),
                  ),
                  const SizedBox(width: AppSpacing.xs),
                  Expanded(child: Text(warning, style: style)),
                ],
              ),
            ),
          ),
      ],
    );
  }
}

/// "¿Es deducible?": shows the suggested value until the user overrides
/// it, says which one it shows, and offers "Usar el valor sugerido" while
/// overridden (`deductible: null`).
class DeductibleSwitch extends StatelessWidget {
  /// The switch of [model].
  const new({required this.model, super.key});

  /// The form whose override this edits.
  final ExpenseFormModel model;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final overridden = model.deductibleOverride != null;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        SwitchListTile(
          focusNode: model.deductible.focusNode,
          contentPadding: EdgeInsets.zero,
          title: Text(model.deductible.label),
          subtitle: Text(
            overridden
                ? l10n.expenseDeductibleUser
                : l10n.expenseDeductibleSuggested,
          ),
          value: model.effectiveDeductible,
          onChanged: (value) => model.deductibleOverride = value,
        ),
        if (overridden)
          Align(
            alignment: Alignment.centerLeft,
            child: TextButton(
              style: TextButton.styleFrom(
                minimumSize: const Size(0, AppSizes.minTarget),
              ),
              onPressed: () => model.deductibleOverride = null,
              child: Text(l10n.expenseDeductibleReset),
            ),
          ),
      ],
    );
  }
}
